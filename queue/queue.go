package queue

import (
	"fmt"
	"slices"
	"sync"

	"github.com/go-redsync/redsync/v4"
	"github.com/go-redsync/redsync/v4/redis/goredis/v9"
	"github.com/redis/go-redis/v9"
)

type Queue struct {
	Name          string
	client        *redis.Client
	rs            *redsync.Redsync
	jobs          []Job
	RetryFailures int
	workers       int
}

type QueueOption struct {
	Redis         *redis.Options
	Workers       int
	RetryFailures int
}

func New(name string, opt *QueueOption) *Queue {
	client := redis.NewClient(opt.Redis)

	pool := goredis.NewPool(client)
	rs := redsync.New(pool)

	return &Queue{
		client:        client,
		Name:          name,
		rs:            rs,
		workers:       opt.Workers,
		RetryFailures: opt.RetryFailures,
	}
}

func (q *Queue) AddJob(id string, data interface{}) {
	job := q.newJob(id, data)
	q.jobs = append(q.jobs, *job)
}

type JobFnc func(job *Job)

func (q *Queue) Process(jobFnc JobFnc) {
	mutex := q.rs.NewMutex(q.Name)
	if err := mutex.Lock(); err != nil {
		panic(err)
	}
	q.Run(jobFnc)
	if ok, err := mutex.Unlock(); !ok || err != nil {
		panic("unlock failed")
	}
}

func (q *Queue) Run(jobFnc JobFnc) {
	execJobs := []*Job{}
	for i := 0; i < len(q.jobs); i++ {
		if q.jobs[i].IsReady() {
			execJobs = append(execJobs, &q.jobs[i])
		}
	}

	for len(execJobs) > 0 {
		min := Min(len(execJobs), q.workers)
		numJobs := execJobs[:min]
		var wg sync.WaitGroup
		for i := 0; i < len(numJobs); i++ {
			job := numJobs[i]
			wg.Add(1)
			go func() {
				defer wg.Done()
				jobFnc(job)
			}()
		}
		wg.Wait()
		_, execJobs = execJobs[0], execJobs[min:]
	}

	q.Retry(jobFnc)
}

func (q *Queue) Retry(jobFnc JobFnc) {
	execJobs := []*Job{}
	// For retry failures
	for i := 0; i < len(q.jobs); i++ {
		if q.jobs[i].Status == DelayedStatus {
			execJobs = append(execJobs, &q.jobs[i])
		}
	}

	for len(execJobs) > 0 {
		min := Min(len(execJobs), q.workers)
		numJobs := execJobs[:min]
		var wg sync.WaitGroup
		for i := 0; i < len(numJobs); i++ {
			job := numJobs[i]
			wg.Add(1)
			go func() {
				defer wg.Done()
				jobFnc(job)
				if job.Status == CompletedStatus || job.Status == FailedStatus {
					_, execJobs = execJobs[0], execJobs[1:]
				}
			}()
		}
		wg.Wait()
	}
}

func (q *Queue) CountJobs(status JobStatus) int {
	count := 0
	for i := 0; i < len(q.jobs); i++ {
		if q.jobs[i].Status == status {
			count++
		}
	}

	return count
}

func (q *Queue) Remove(key string) {
	findIdx := slices.IndexFunc(q.jobs, func(j Job) bool { return j.Id == key })
	if findIdx != -1 {
		fmt.Print(findIdx)
		q.jobs = append(q.jobs[:findIdx], q.jobs[findIdx+1:]...)
	}
}

func Min(a int, b int) int {
	if a < b {
		return a
	}
	return b
}
