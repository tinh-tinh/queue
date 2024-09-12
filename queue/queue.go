package queue

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"strconv"
	"sync"
	"time"

	"github.com/go-redsync/redsync/v4"
	"github.com/go-redsync/redsync/v4/redis/goredis/v9"
	"github.com/redis/go-redis/v9"
	"github.com/robfig/cron/v3"
)

type Queue struct {
	Name          string
	client        *redis.Client
	rs            *redsync.Redsync
	jobs          []Job
	RetryFailures int
	workers       int
	limiter       *RateLimiter
	ctx           context.Context
	scheduler     *cron.Cron
	cronPattern   string
}

type RateLimiter struct {
	Max      int
	Duration time.Duration
}

type QueueOption struct {
	Workers       int
	RetryFailures int
	Limiter       *RateLimiter
	Pattern       string
}

func New(name string, opt *QueueOption, client *redis.Client) *Queue {
	pool := goredis.NewPool(client)
	rs := redsync.New(pool)

	queue := &Queue{
		client:        client,
		Name:          name,
		rs:            rs,
		workers:       opt.Workers,
		RetryFailures: opt.RetryFailures,
		limiter:       opt.Limiter,
		ctx:           context.Background(),
	}

	if opt.Pattern != "" {
		queue.scheduler = cron.New()
		queue.cronPattern = opt.Pattern
	}

	return queue
}

func (q *Queue) AddJob(id string, data interface{}) {
	var job *Job
	if q.IsLimit() {
		fmt.Printf("Add job %s to delay\n", id)
		job = q.delayJob(id, data)
	} else {
		fmt.Printf("Add job %s to waiting\n", id)
		job = q.newJob(id, data)
	}
	q.jobs = append(q.jobs, *job)
}

type JobFnc func(job *Job)

func (q *Queue) Process(jobFnc JobFnc) {
	if q.scheduler == nil {
		mutex := q.rs.NewMutex(q.Name)
		if err := mutex.Lock(); err != nil {
			panic(err)
		}
		q.Run(jobFnc)
		if ok, err := mutex.Unlock(); !ok || err != nil {
			panic("unlock failed")
		}
	} else {
		fmt.Println(q.cronPattern)
		// q.scheduler.AddFunc(q.cronPattern, func() {
		// 	// mutex := q.rs.NewMutex(q.Name)
		// 	// if err := mutex.Lock(); err != nil {
		// 	// 	panic(err)
		// 	// }
		// 	q.Run(jobFnc)
		// 	// if ok, err := mutex.Unlock(); !ok || err != nil {
		// 	// 	panic("unlock failed")
		// 	// }
		// })
		q.scheduler.AddFunc(q.cronPattern, func() { fmt.Println("Every second") })
		q.scheduler.Start()
	}
}

func (q *Queue) Run(jobFnc JobFnc) {
	fmt.Printf("Running on %s\n", time.Now().String())
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
				if job.IsFinished() {
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

func (q *Queue) IsLimit() bool {
	if q.limiter == nil {
		return false
	}
	client := q.client
	attemps, _ := client.Get(q.ctx, q.Name).Result()
	attempNum, _ := strconv.Atoi(attemps)
	if attemps != "" && attempNum >= q.limiter.Max {
		return true
	} else {
		value, err := client.Incr(q.ctx, q.Name).Result()
		if err != nil {
			panic(errors.New("fail to incr data"))
		}
		if value == 1 {
			client.Expire(q.ctx, q.Name, q.limiter.Duration)
		}
		return false
	}
}
