package queue

import (
	"context"
	"log"
	"slices"
	"sort"
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
	Name        string
	client      *redis.Client
	mutex       *redsync.Mutex
	jobFnc      JobFnc
	jobs        []Job
	ctx         context.Context
	scheduler   *cron.Cron
	cronPattern string
	running     bool
	config      Options
}

type RateLimiter struct {
	Max      int
	Duration time.Duration
}

type Options struct {
	Connect          *redis.Options
	Workers          int
	RetryFailures    int
	Limiter          *RateLimiter
	Pattern          string
	Logger           LoggerType
	RemoveOnComplete bool
	RemoveOnFail     bool
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
func New(name string, opt *Options) *Queue {
	client := redis.NewClient(opt.Connect)
	pool := goredis.NewPool(client)
	rs := redsync.New(pool)

	queue := &Queue{
		client:  client,
		Name:    name,
		mutex:   rs.NewMutex(name),
		ctx:     context.Background(),
		running: true,
		config:  *opt,
	}

	if opt.Logger == "" {
		queue.config.Logger = LoggerDefault
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
func (q *Queue) AddJob(opt AddJobOptions) {
	var job *Job
	if q.IsLimit() {
		q.formatLog(LoggerInfo, "Add job %s to delay\n", opt.Id)
		job = q.delayJob(opt)
	} else {
		q.formatLog(LoggerInfo, "Add job %s to waiting\n", opt.Id)
		job = q.newJob(opt)
	}
	q.jobs = append(q.jobs, *job)
	sort.SliceStable(q.jobs, func(i, j int) bool { return q.jobs[i].Priority > q.jobs[j].Priority })
	q.Run()
}

type AddJobOptions struct {
	Id       string
	Data     interface{}
	Priority int
}

// BulkAddJob adds multiple jobs to the queue at once. If the queue is currently
// rate limited, the jobs are delayed. Otherwise, the jobs are added to the
// waiting list and the queue is run.
func (q *Queue) BulkAddJob(options []AddJobOptions) {
	sort.SliceStable(options, func(i, j int) bool { return options[i].Priority > options[j].Priority })
	for _, option := range options {
		var job *Job
		if q.IsLimit() {
			q.formatLog(LoggerInfo, "Add job %s to delay\n", option.Id)
			job = q.delayJob(option)
		} else {
			q.formatLog(LoggerInfo, "Add job %s to waiting\n", option.Id)
			job = q.newJob(option)
		}
		q.jobs = append(q.jobs, *job)
	}
	sort.SliceStable(q.jobs, func(i, j int) bool { return q.jobs[i].Priority > q.jobs[j].Priority })
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
			q.formatLog(LoggerFatal, "failed to add job: %v\n", err)
		}
		q.scheduler.Start()
	}
}

// Run runs all ready jobs in the queue. It locks the mutex, runs all ready jobs
// in parallel, and then unlocks the mutex. If the queue has a scheduler, it
// will be started with the given cron pattern. Otherwise, the callback is simply
// stored.
func (q *Queue) Run() {
	if !q.running {
		q.formatLog(LoggerInfo, "Queue is not running")
		return
	}
	// Lock the mutex
	if err := q.mutex.Lock(); err != nil {
		q.formatLog(LoggerFatal, "Error when lock mutex: %v\n", err)
		return
	}
	execJobs := []*Job{}
	for i := range q.jobs {
		if q.jobs[i].IsReady() {
			execJobs = append(execJobs, &q.jobs[i])
		}
	}

	for len(execJobs) > 0 {
		min := Min(len(execJobs), q.config.Workers)
		numJobs := execJobs[:min]
		var wg sync.WaitGroup
		for i := range numJobs {
			job := numJobs[i]
			wg.Add(1)
			go func() {
				defer wg.Done()
				q.jobFnc(job)
			}()
		}
		wg.Wait()
		execJobs = execJobs[min:]
		q.RemoveCompleted()
		q.RemoveFailed()
	}

	q.Retry()
	// Unlock the mutex
	if ok, err := q.mutex.Unlock(); !ok || err != nil {
		q.formatLog(LoggerFatal, "Error when unlock mutex: %v\n", err)
	}
}

// Retry processes all jobs that are in the DelayedStatus. It locks the mutex,
// collects all delayed jobs, and then processes them concurrently up to the
// number of available workers. After processing, it checks if the job is finished
// and removes it from the list of jobs to retry. Finally, it unlocks the mutex.

func (q *Queue) Retry() {
	execJobs := []*Job{}
	// For retry failures
	for i := range q.jobs {
		if q.jobs[i].Status == DelayedStatus {
			execJobs = append(execJobs, &q.jobs[i])
		}
	}

	for len(execJobs) > 0 {
		min := Min(len(execJobs), q.config.Workers)
		numJobs := execJobs[:min]
		var wg sync.WaitGroup
		var finishedJob []string
		for i := range numJobs {
			job := numJobs[i]
			wg.Add(1)
			go func() {
				defer wg.Done()
				q.jobFnc(job)
				if job.IsFinished() {
					finishedJob = append(finishedJob, job.Id)
				}
			}()
		}
		wg.Wait()
		if len(finishedJob) > 0 {
			for _, id := range finishedJob {
				if len(execJobs) == 1 && execJobs[0].Id == id {
					execJobs = []*Job{}
					break
				}
				idx := slices.IndexFunc(execJobs, func(j *Job) bool { return j.Id == id })
				if idx != -1 {
					execJobs = append(execJobs[:idx], execJobs[idx+1:]...)
				} else {
					break
				}
			}
		}
		q.RemoveFailed()
	}
}

// CountJobs returns the number of jobs in the queue that have the given status.
//
// This can be used to monitor the queue, and to test the queue's behavior.
func (q *Queue) CountJobs(status JobStatus) int {
	count := 0
	for i := range q.jobs {
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
		q.jobs = slices.Delete(q.jobs, findIdx, findIdx+1)
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
	if q.config.Limiter == nil {
		return false
	}
	client := q.client
	attemps, _ := client.Get(q.ctx, q.Name).Result()
	attempNum, _ := strconv.Atoi(attemps)
	if attemps != "" && attempNum >= q.config.Limiter.Max {
		return true
	} else {
		value, err := client.Incr(q.ctx, q.Name).Result()
		if err != nil {
			q.formatLog(LoggerPanic, "Error when count redis: %v\n", err)
		}
		if value == 1 {
			client.Expire(q.ctx, q.Name, q.config.Limiter.Duration)
		}
		return false
	}
}

// Pause stops the queue from running. When paused, the queue will not accept new
// jobs and will not run any jobs in the queue. It will resume when Resume is
// called.
func (q *Queue) Pause() {
	q.running = false
}

// Resume resumes the queue from a paused state. When resumed, the queue will
// accept new jobs and run any jobs in the queue.
func (q *Queue) Resume() {
	q.running = true
	q.Run()
}

func (q *Queue) RemoveCompleted() {
	if q.config.RemoveOnComplete {
		q.jobs = Remove(q.jobs, func(j Job) bool {
			return j.Status == CompletedStatus
		})
	}
}

func (q *Queue) RemoveFailed() {
	if q.config.RemoveOnFail {
		q.jobs = Remove(q.jobs, func(j Job) bool {
			return j.Status == FailedStatus
		})
	}
}

func (q *Queue) formatLog(logType LoggerType, format string, v ...any) {
	if q.config.Logger == LoggerDisabled {
		return
	} else if q.config.Logger == LoggerDefault {
		q.log(logType, format, v...)
	} else if q.config.Logger == logType {
		q.log(logType, format, v...)
	}
}

func (q *Queue) log(logType LoggerType, format string, v ...any) {
	if logType == LoggerInfo {
		log.Printf(format, v...)
	} else if logType == LoggerFatal {
		log.Fatalf(format, v...)
	} else if logType == LoggerPanic {
		log.Panicf(format, v...)
	}
}
