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
		RetryFailures: queue.RetryFailures,
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
		RetryFailures: queue.RetryFailures,
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
	fmt.Printf("⌛ Running job %s progress\n", job.Id)
	println()
	defer func() {
		if r := recover(); r != nil {
			job.FailedReason = fmt.Sprintf("%v", r)
			job.Status = FailedStatus
			// Store error
			client := job.queue.client
			_, err := client.HSet(context.Background(), fmt.Sprintf("%sstore", job.queue.Name), job.Id, job.FailedReason).Result()
			if err != nil {
				fmt.Printf("Failed to store error: %s\n", err.Error())
			}
			if job.RetryFailures > 0 {
				job.Status = DelayedStatus
				job.RetryFailures--
				fmt.Printf("Add job %s for retry (%d remains) 🕛\n", job.Id, job.RetryFailures)
				println()
			} else {
				fmt.Printf("Failed job %s ❌\n", job.Id)
				println()
			}
		}
	}()
	err := cb()
	if err == nil {
		job.FinishedOn = time.Now()
		job.Status = CompletedStatus
		fmt.Printf("Job %s done ✅ in %dms\n", job.Id, job.FinishedOn.Sub(job.ProcessedOn).Milliseconds())
		println()
	} else {
		job.FailedReason = err.Error()
		job.Status = FailedStatus
		// Store error
		client := job.queue.client
		_, err := client.HSet(context.Background(), fmt.Sprintf("%sstore", job.queue.Name), job.Id, job.FailedReason).Result()
		if err != nil {
			fmt.Printf("Failed to store error: %s\n", err.Error())
		}
		if job.RetryFailures > 0 {
			job.Status = DelayedStatus
			job.RetryFailures--
			fmt.Printf("Add job %s for retry (%d remains) 🕛\n", job.Id, job.RetryFailures)
			println()
		} else {
			fmt.Printf("Failed job %s ❌\n", job.Id)
			println()
		}
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
