package queue_test

import (
	"sync"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/require"
	"github.com/tinh-tinh/queue/v2"
)

func Test_ScheduleJob(t *testing.T) {
	schedulerQueue := queue.New("scheduler_test", &queue.Options{
		Connect: &redis.Options{
			Addr:     "localhost:6379",
			Password: "",
			DB:       0,
		},
		Workers:          3,
		RetryFailures:    0,
		Pattern:          "@every 1s", // Enable scheduler
		ScheduleInterval: 1 * time.Second,
	})

	// Track processed jobs
	processedJobs := make(map[string]bool)
	var mu sync.Mutex

	schedulerQueue.Process(func(job *queue.Job) {
		job.Process(func() error {
			mu.Lock()
			processedJobs[job.Id] = true
			mu.Unlock()
			return nil
		})
	})

	// Schedule a job to run 2 seconds from now
	runAt := time.Now().Add(2 * time.Second)
	err := schedulerQueue.ScheduleJob("scheduled_job_1", runAt)
	require.Nil(t, err)

	// Verify job is in scheduled set
	scheduledJobs, err := schedulerQueue.GetScheduledJobs()
	require.Nil(t, err)
	require.Equal(t, 1, len(scheduledJobs))
	require.Equal(t, "scheduled_job_1", scheduledJobs[0].JobId)

	// Wait for job to be processed (2s + 1s buffer)
	time.Sleep(3 * time.Second)

	// Verify job was processed
	mu.Lock()
	require.True(t, processedJobs["scheduled_job_1"])
	mu.Unlock()

	// Verify job is no longer in scheduled set
	scheduledJobs, err = schedulerQueue.GetScheduledJobs()
	require.Nil(t, err)
	require.Equal(t, 0, len(scheduledJobs))
}

func Test_RemoveScheduledJob(t *testing.T) {
	schedulerQueue := queue.New("remove_scheduled_test", &queue.Options{
		Connect: &redis.Options{
			Addr:     "localhost:6379",
			Password: "",
			DB:       0,
		},
		Workers:          3,
		RetryFailures:    0,
		Pattern:          "@every 1s",
		ScheduleInterval: 1 * time.Second,
	})

	// Schedule a job for 5 seconds from now
	runAt := time.Now().Add(5 * time.Second)
	err := schedulerQueue.ScheduleJob("job_to_remove", runAt)
	require.Nil(t, err)

	// Verify job is scheduled
	scheduledJobs, err := schedulerQueue.GetScheduledJobs()
	require.Nil(t, err)
	require.Equal(t, 1, len(scheduledJobs))

	// Remove the scheduled job
	err = schedulerQueue.RemoveScheduledJob("job_to_remove")
	require.Nil(t, err)

	// Verify job is no longer scheduled
	scheduledJobs, err = schedulerQueue.GetScheduledJobs()
	require.Nil(t, err)
	require.Equal(t, 0, len(scheduledJobs))
}

func Test_PauseScheduler(t *testing.T) {
	pauseQueue := queue.New("pause_scheduler_test", &queue.Options{
		Connect: &redis.Options{
			Addr:     "localhost:6379",
			Password: "",
			DB:       0,
		},
		Workers:          3,
		RetryFailures:    0,
		Pattern:          "@every 1s",
		ScheduleInterval: 1 * time.Second,
	})

	// Track processed jobs
	processedJobs := make(map[string]bool)
	var mu sync.Mutex

	pauseQueue.Process(func(job *queue.Job) {
		job.Process(func() error {
			mu.Lock()
			processedJobs[job.Id] = true
			mu.Unlock()
			return nil
		})
	})

	// Schedule a job to run 2 seconds from now
	runAt := time.Now().Add(2 * time.Second)
	err := pauseQueue.ScheduleJob("paused_job", runAt)
	require.Nil(t, err)

	// Pause the queue immediately
	pauseQueue.Pause()

	// Wait for when the job should have been processed
	time.Sleep(3 * time.Second)

	// Verify job was NOT processed (scheduler stopped)
	mu.Lock()
	require.False(t, processedJobs["paused_job"])
	mu.Unlock()

	// Verify job is still in scheduled set
	scheduledJobs, err := pauseQueue.GetScheduledJobs()
	require.Nil(t, err)
	require.Equal(t, 1, len(scheduledJobs))
	require.Equal(t, "paused_job", scheduledJobs[0].JobId)
}

func Test_ResumeScheduler(t *testing.T) {
	resumeQueue := queue.New("resume_scheduler_test", &queue.Options{
		Connect: &redis.Options{
			Addr:     "localhost:6379",
			Password: "",
			DB:       0,
		},
		Workers:          3,
		RetryFailures:    0,
		Pattern:          "@every 1s",
		ScheduleInterval: 1 * time.Second,
	})

	// Track processed jobs
	processedJobs := make(map[string]bool)
	var mu sync.Mutex

	resumeQueue.Process(func(job *queue.Job) {
		job.Process(func() error {
			mu.Lock()
			processedJobs[job.Id] = true
			mu.Unlock()
			return nil
		})
	})

	// Pause the queue
	resumeQueue.Pause()

	// Schedule a job to run 3 seconds from now
	runAt := time.Now().Add(3 * time.Second)
	err := resumeQueue.ScheduleJob("resume_job", runAt)
	require.Nil(t, err)

	// Resume the queue
	resumeQueue.Resume()

	// Wait for job to be processed (3s + 2s buffer)
	time.Sleep(5 * time.Second)

	// Verify job was processed (scheduler restarted)
	mu.Lock()
	require.True(t, processedJobs["resume_job"])
	mu.Unlock()

	// Verify job is no longer in scheduled set
	scheduledJobs, err := resumeQueue.GetScheduledJobs()
	require.Nil(t, err)
	require.Equal(t, 0, len(scheduledJobs))
}

func Test_PauseResumeMultipleScheduledJobs(t *testing.T) {
	multiQueue := queue.New("multi_pause_resume_test", &queue.Options{
		Connect: &redis.Options{
			Addr:     "localhost:6379",
			Password: "",
			DB:       0,
		},
		Workers:          3,
		RetryFailures:    0,
		Pattern:          "@every 1s",
		ScheduleInterval: 1 * time.Second,
	})

	// Track processed jobs
	processedJobs := make(map[string]bool)
	var mu sync.Mutex

	multiQueue.Process(func(job *queue.Job) {
		job.Process(func() error {
			mu.Lock()
			processedJobs[job.Id] = true
			mu.Unlock()
			return nil
		})
	})

	// Schedule multiple jobs
	runAt1 := time.Now().Add(2 * time.Second)
	runAt2 := time.Now().Add(3 * time.Second)
	runAt3 := time.Now().Add(4 * time.Second)

	err := multiQueue.ScheduleJob("job1", runAt1)
	require.Nil(t, err)
	err = multiQueue.ScheduleJob("job2", runAt2)
	require.Nil(t, err)
	err = multiQueue.ScheduleJob("job3", runAt3)
	require.Nil(t, err)

	// Verify all jobs are scheduled
	scheduledJobs, err := multiQueue.GetScheduledJobs()
	require.Nil(t, err)
	require.Equal(t, 3, len(scheduledJobs))

	// Pause after 2.5 seconds (job1 should have been processed)
	time.Sleep(2500 * time.Millisecond)
	multiQueue.Pause()

	// Wait a bit more
	time.Sleep(2 * time.Second)

	// Verify only job1 was processed
	mu.Lock()
	require.True(t, processedJobs["job1"])
	require.False(t, processedJobs["job2"])
	require.False(t, processedJobs["job3"])
	mu.Unlock()

	// Resume the queue
	multiQueue.Resume()

	// Wait for remaining jobs to be processed
	time.Sleep(2 * time.Second)

	// Verify all jobs were eventually processed
	mu.Lock()
	require.True(t, processedJobs["job1"])
	require.True(t, processedJobs["job2"])
	require.True(t, processedJobs["job3"])
	mu.Unlock()

	// Verify all jobs are removed from scheduled set
	scheduledJobs, err = multiQueue.GetScheduledJobs()
	require.Nil(t, err)
	require.Equal(t, 0, len(scheduledJobs))
}
