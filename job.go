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

func (queue *Queue) newJob(id string, data interface{}) *Job {
	job := &Job{
		Id:            id,
		Data:          data,
		Status:        WaitStatus,
		Stacktrace:    []string{},
		queue:         queue,
		RetryFailures: queue.RetryFailures,
	}

	return job
}

func (queue *Queue) delayJob(id string, data interface{}) *Job {
	job := &Job{
		Id:            id,
		Data:          data,
		Status:        DelayedStatus,
		Stacktrace:    []string{},
		queue:         queue,
		RetryFailures: queue.RetryFailures,
	}

	return job
}

type Callback func() error

func (job *Job) Process(cb Callback) {
	job.Status = ActiveStatus
	job.ProcessedOn = time.Now()
	fmt.Printf("⌛ Running job %s progress\n", job.Id)
	println()
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
		client.Set(context.Background(), job.Id, job.FailedReason, time.Hour)
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

func (job *Job) IsReady() bool {
	if job.queue.scheduler == nil {
		return job.Status == WaitStatus || job.Status == ActiveStatus
	}
	return true
}

func (job *Job) IsFinished() bool {
	return job.Status == FailedStatus || job.Status == CompletedStatus
}
