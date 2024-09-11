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

func New(name string, opt *redis.Options) *Queue {
	client := redis.NewClient(opt)

	pool := goredis.NewPool(client)
	rs := redsync.New(pool)

	return &Queue{
		client:  client,
		Name:    name,
		rs:      rs,
		workers: 3,
	}
}

func (q *Queue) AddJob(id string, data interface{}, opts ...AddJobOptions) {
	if len(opts) == 0 {
		opts = append(opts, AddJobOptions{
			RemoveOnComplete: true,
			RemoveOnFail:     true,
		})
	}
	job := q.newJob(id, data, opts[0])
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
	for len(q.jobs) > 0 {
		min := Min(len(q.jobs), q.workers)
		numJobs := q.jobs[:min]
		var wg sync.WaitGroup
		for i := 0; i < len(numJobs); i++ {
			job := q.jobs[i]
			if job.IsReady() {
				wg.Add(1)
				go func() {
					defer wg.Done()
					jobFnc(&job)
				}()
			}

			if q.CountJobs(WaitStatus) == 0 {
				break
			}
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
