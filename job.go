package queue

import (
	"context"
	"fmt"
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
	job.queue.formatLog(LoggerInfo, "Running job %s progress\n\n", job.Id)
	defer func() {
		if r := recover(); r != nil {
			failedReason := fmt.Sprintf("%v", r)
			job.HandlerError(failedReason)
		}
	}()
	err := cb()
	if err == nil {
		job.FinishedOn = time.Now()
		job.Status = CompletedStatus
		job.queue.formatLog(LoggerInfo, "Job %s done in %dms\n\n", job.Id, job.FinishedOn.Sub(job.ProcessedOn).Milliseconds())
	} else {
		job.HandlerError(err.Error())
	}
}

func (job *Job) HandlerError(reasonError string) {
	job.FailedReason = reasonError
	job.Status = FailedStatus
	// Store error
	client := job.queue.client
	_, err := client.HSet(context.Background(), job.queue.Name, job.Id, job.FailedReason).Result()
	if err != nil {
		job.queue.formatLog(LoggerInfo, "Failed to store error: %s\n\n", err.Error())
	}
	if job.RetryFailures > 0 {
		job.Status = DelayedStatus
		job.RetryFailures--
		job.queue.formatLog(LoggerInfo, "Add job %s for retry (%d remains) \n\n", job.Id, job.RetryFailures)
	} else {
		job.queue.formatLog(LoggerInfo, "Failed job %s \n\n", job.Id)
	}
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
