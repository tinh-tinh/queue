package queue

import (
	"context"
	"fmt"
	"strings"
	"time"
)

type JobStatus string

const (
	WaitStatus      JobStatus = "wait"
	DelayedStatus   JobStatus = "delayed"
	ActiveStatus    JobStatus = "active"
	CompletedStatus JobStatus = "completed"
	FailedStatus    JobStatus = "failed"
)

type Job struct {
	Id            string
	Data          interface{}
	Priority      int
	Status        JobStatus
	queue         *Queue
	ProcessedOn   time.Time
	FinishedOn    time.Time
	Stacktrace    []string
	FailedReason  string
	RetryFailures int
}

// newJob creates a new job with the given id and data. It sets the status to WaitStatus
// and sets the RetryFailures to the RetryFailures of the Queue.
func (queue *Queue) newJob(opt AddJobOptions) *Job {
	job := &Job{
		Id:            opt.Id,
		Data:          opt.Data,
		Priority:      opt.Priority,
		Status:        WaitStatus,
		Stacktrace:    []string{},
		queue:         queue,
		RetryFailures: queue.config.RetryFailures,
	}

	return job
}

// delayJob creates a new job with the given id and data, and sets its status to
// DelayedStatus.
func (queue *Queue) delayJob(opt AddJobOptions) *Job {
	job := &Job{
		Id:            opt.Id,
		Data:          opt.Data,
		Priority:      opt.Priority,
		Status:        DelayedStatus,
		Stacktrace:    []string{},
		queue:         queue,
		RetryFailures: queue.config.RetryFailures,
	}

	return job
}

type Callback func() error

// Process runs the given callback and updates the job's status accordingly. It
// also measures and logs the execution time. If the callback returns an error,
// the job is either retried or marked as failed.
func (job *Job) Process(cb Callback) {
	job.Status = ActiveStatus
	job.ProcessedOn = time.Now()
	job.queue.formatLog(LoggerInfo, "Running job %s progress", job.Id)

	defer func() {
		if r := recover(); r != nil {
			failedReason := fmt.Sprintf("%v", r)
			job.HandlerError(failedReason)
		}
	}()

	err := cb()
	if err != nil {
		job.HandlerError(err.Error())
		return
	}

	job.FinishedOn = time.Now()
	job.Status = CompletedStatus
	job.queue.formatLog(LoggerInfo, "Job %s done in %dms", job.Id, job.FinishedOn.Sub(job.ProcessedOn).Milliseconds())
}

func (job *Job) HandlerError(reasonError string) {
	job.FailedReason = reasonError
	job.Status = FailedStatus

	// Store error
	if job.RetryFailures <= 0 {
		client := job.queue.client
		key := job.getKey()
		_, err := client.Set(context.Background(), key, job.FailedReason, 0).Result()
		if err != nil {
			job.queue.formatLog(LoggerFatal, "Failed to store error: %s", err.Error())
		}
		job.queue.formatLog(LoggerError, "Failed job %s ", job.Id)
		return
	}

	job.Status = DelayedStatus
	job.RetryFailures--
	job.queue.formatLog(LoggerWarn, "Add job %s for retry (%d remains) ", job.Id, job.RetryFailures)
}

// IsReady returns true if the job is ready to be processed. If the job uses a
// scheduler, it will always be ready. Otherwise, the job is ready if it is
// waiting or active.
func (job *Job) IsReady() bool {
	if job.queue.scheduler == nil {
		return job.Status == WaitStatus || job.Status == ActiveStatus
	}
	return true
}

// IsFinished returns true if the job has finished, either successfully or with an error.
func (job *Job) IsFinished() bool {
	return job.Status == FailedStatus || job.Status == CompletedStatus
}

func (job *Job) getKey() string {
	if job.queue.config.Prefix != "" {
		prefix := job.queue.config.Prefix
		return fmt.Sprintf("%s:%s", strings.ToLower(prefix+job.queue.Name), job.Id)
	}
	return fmt.Sprintf("%s:%s", strings.ToLower(job.queue.Name), job.Id)
}
