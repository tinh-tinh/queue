package queue

import (
	"context"
	"errors"
	"fmt"
	"log"
	"slices"
	"strconv"
	"sync"
	"time"

	"github.com/go-redsync/redsync/v4"
	"github.com/go-redsync/redsync/v4/redis/goredis/v9"
	"github.com/redis/go-redis/v9"
	"github.com/robfig/cron/v3"
)

type JobFnc func(job *Job)

type Queue struct {
	Name          string
	client        *redis.Client
	mutex         *redsync.Mutex
	jobFnc        JobFnc
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
	Connect       *redis.Options
	Workers       int
	RetryFailures int
	Limiter       *RateLimiter
	Pattern       string
}

// New creates a new queue with the given name and options. The name is used to
// identify the queue in Redis, and the options are used to configure the queue
// behavior. The options are as follows:
//
// - Connect: the Redis connection options
// - Workers: the number of workers to run concurrently
// - RetryFailures: the number of times to retry a failed job
// - Limiter: the rate limiter options
// - Pattern: the cron pattern to use for scheduling jobs
//
// The returned queue is ready to use.
func New(name string, opt *QueueOption) *Queue {
	client := redis.NewClient(opt.Connect)
	pool := goredis.NewPool(client)
	rs := redsync.New(pool)

	queue := &Queue{
		client:        client,
		Name:          name,
		mutex:         rs.NewMutex(name),
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

// AddJob adds a new job to the queue. If the queue is currently rate limited, the
// job is delayed. Otherwise, the job is added to the waiting list and the queue
// is run.
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
	q.Run()
}

type AddJobOptions struct {
	Id   string
	Data interface{}
}

// BulkAddJob adds multiple jobs to the queue at once. If the queue is currently
// rate limited, the jobs are delayed. Otherwise, the jobs are added to the
// waiting list and the queue is run.
func (q *Queue) BulkAddJob(options []AddJobOptions) {
	for _, option := range options {
		var job *Job
		if q.IsLimit() {
			fmt.Printf("Add job %s to delay\n", option.Id)
			job = q.delayJob(option.Id, option.Data)
		} else {
			fmt.Printf("Add job %s to waiting\n", option.Id)
			job = q.newJob(option.Id, option.Data)
		}
		q.jobs = append(q.jobs, *job)
	}
	q.Run()
}

// Process sets the callback for the queue to process jobs. If the queue has a
// scheduler, it will be started with the given cron pattern. Otherwise, the
// callback is simply stored.
func (q *Queue) Process(jobFnc JobFnc) {
	q.jobFnc = jobFnc
	if q.scheduler != nil {
		_, err := q.scheduler.AddFunc(q.cronPattern, func() { q.Run() })
		if err != nil {
			log.Fatalf("failed to add job: %v\n", err)
		}
		q.scheduler.Start()
	}
}

// Run runs all ready jobs in the queue. It locks the mutex, runs all ready jobs
// in parallel, and then unlocks the mutex. If the queue has a scheduler, it
// will be started with the given cron pattern. Otherwise, the callback is simply
// stored.
func (q *Queue) Run() {
	// Lock the mutex
	// if err := q.mutex.Lock(); err != nil {
	// 	fmt.Println(err)
	// }
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
				q.jobFnc(job)
			}()
		}
		wg.Wait()
		execJobs = execJobs[min:]
	}

	// Unlock the mutex
	// if ok, err := q.mutex.Unlock(); !ok || err != nil {
	// 	fmt.Println(err)
	// }

	q.Retry()
}

// Retry processes all jobs that are in the DelayedStatus. It locks the mutex,
// collects all delayed jobs, and then processes them concurrently up to the
// number of available workers. After processing, it checks if the job is finished
// and removes it from the list of jobs to retry. Finally, it unlocks the mutex.

func (q *Queue) Retry() {
	// Lock the mutex
	// if err := q.mutex.Lock(); err != nil {
	// 	panic(err)
	// }

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
				q.jobFnc(job)
				if job.IsFinished() {
					execJobs = append(execJobs[:i], execJobs[i+1:]...)
				}
			}()
		}
		wg.Wait()
	}

	// Unlock the mutex
	// if ok, err := q.mutex.Unlock(); !ok || err != nil {
	// 	panic("unlock failed")
	// }
}

// CountJobs returns the number of jobs in the queue that have the given status.
//
// This can be used to monitor the queue, and to test the queue's behavior.
func (q *Queue) CountJobs(status JobStatus) int {
	count := 0
	for i := 0; i < len(q.jobs); i++ {
		if q.jobs[i].Status == status {
			count++
		}
	}

	return count
}

// Remove removes the job with the given key from the queue. It uses a linear
// search, so it has a time complexity of O(n), where n is the number of jobs in
// the queue.
func (q *Queue) Remove(key string) {
	findIdx := slices.IndexFunc(q.jobs, func(j Job) bool { return j.Id == key })
	if findIdx != -1 {
		q.jobs = append(q.jobs[:findIdx], q.jobs[findIdx+1:]...)
	}
}

func Min(a int, b int) int {
	if a < b {
		return a
	}
	return b
}

// IsLimit returns true if the number of jobs in the queue has reached the
// maximum value set in the RateLimiter. It checks the current value of the
// counter in Redis and returns true if it is greater than or equal to the
// maximum value. If the counter does not exist or is less than the maximum,
// it increments the counter and returns false. If the increment fails, it
// panics.
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
