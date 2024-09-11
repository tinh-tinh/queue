package queue

import (
	"context"
	"time"

	"github.com/sirupsen/logrus"
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
	Id               string
	Data             interface{}
	Priority         int
	Status           JobStatus
	queue            *Queue
	ProcessedOn      time.Time
	FinishedOn       time.Time
	Stacktrace       []string
	FailedReason     string
	RetryFailures    int
	RemoveOnComplete bool
	RemoveOnFail     bool
}

type AddJobOptions struct {
	RemoveOnComplete bool
	RemoveOnFail     bool
}

func (queue *Queue) newJob(id string, data interface{}, opt AddJobOptions) *Job {
	job := &Job{
		Id:               id,
		Data:             data,
		Status:           WaitStatus,
		Stacktrace:       []string{},
		queue:            queue,
		RetryFailures:    queue.RetryFailures,
		RemoveOnComplete: opt.RemoveOnComplete,
		RemoveOnFail:     opt.RemoveOnFail,
	}

	return job
}

type Callback func() error

func (job *Job) Process(cb Callback) {
	job.Status = ActiveStatus
	job.ProcessedOn = time.Now()
	logrus.Infof("⌛ Running job %s progress", job.Id)
	println()
	err := cb()
	if err == nil {
		job.FinishedOn = time.Now()
		job.Status = CompletedStatus
		logrus.Infof("Job %s done ✅ in %dms", job.Id, job.FinishedOn.Sub(job.ProcessedOn).Milliseconds())
		println()
		if job.RemoveOnComplete {
			job.queue.Remove(job.Id)
		}
	} else {
		job.FailedReason = err.Error()
		job.Status = FailedStatus
		// Store error
		client := job.queue.client
		client.Set(context.Background(), job.Id, job.FailedReason, time.Hour)
		if job.RetryFailures > 0 {
			job.Status = DelayedStatus
			job.RetryFailures--
			logrus.Warnf("Add job %s for retry (%d remains) 🕛", job.Id, job.RetryFailures)
			println()
		} else {
			logrus.Errorf("Failed job %s ❌", job.Id)
			println()
			if job.RemoveOnFail {
				job.queue.Remove(job.Id)
			}
		}
	}
}

func (job *Job) IsReady() bool {
	return job.Status == WaitStatus || job.Status == ActiveStatus
}
