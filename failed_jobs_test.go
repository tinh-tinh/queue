package queue_test

import (
	"errors"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/require"
	"github.com/tinh-tinh/queue/v2"
)

func Test_GetFailedJobs(t *testing.T) {
	failedQueue := queue.New("failed_jobs_test", &queue.Options{
		Connect: &redis.Options{
			Addr:     "localhost:6379",
			Password: "",
			DB:       0,
		},
		Workers:       3,
		RetryFailures: 0, // No retries, so jobs fail immediately
	})

	// Clear any existing failed jobs first
	err := failedQueue.ClearFailedJobs()
	require.Nil(t, err)

	failedQueue.Process(func(job *queue.Job) {
		job.Process(func() error {
			// All jobs will fail
			return errors.New("intentional failure for test")
		})
	})

	// Add multiple jobs that will fail
	failedQueue.AddJob(queue.AddJobOptions{
		Id:   "fail1",
		Data: "value 1",
	})
	failedQueue.AddJob(queue.AddJobOptions{
		Id:   "fail2",
		Data: "value 2",
	})
	failedQueue.AddJob(queue.AddJobOptions{
		Id:   "fail3",
		Data: "value 3",
	})

	// Wait a bit for jobs to be processed
	time.Sleep(500 * time.Millisecond)

	// Retrieve failed jobs
	failedJobs, err := failedQueue.GetFailedJobs()
	require.Nil(t, err)
	require.Equal(t, 3, len(failedJobs))

	// Verify job IDs are present
	jobIds := make(map[string]bool)
	for _, job := range failedJobs {
		jobIds[job.Id] = true
		require.NotEmpty(t, job.FailedReason)
		require.Contains(t, job.FailedReason, "intentional failure for test")
		require.Equal(t, queue.FailedStatus, job.Status)
	}
	require.True(t, jobIds["fail1"])
	require.True(t, jobIds["fail2"])
	require.True(t, jobIds["fail3"])

	// Clean up
	err = failedQueue.ClearFailedJobs()
	require.Nil(t, err)
}

func Test_GetFailedJob(t *testing.T) {
	singleFailQueue := queue.New("single_fail_test", &queue.Options{
		Connect: &redis.Options{
			Addr:     "localhost:6379",
			Password: "",
			DB:       0,
		},
		Workers:       3,
		RetryFailures: 0,
	})

	// Clear any existing failed jobs first
	err := singleFailQueue.ClearFailedJobs()
	require.Nil(t, err)

	singleFailQueue.Process(func(job *queue.Job) {
		job.Process(func() error {
			return errors.New("specific error for job " + job.Id)
		})
	})

	// Add a job that will fail
	singleFailQueue.AddJob(queue.AddJobOptions{
		Id:   "specific_fail",
		Data: "test data",
	})

	// Wait for job to be processed
	time.Sleep(500 * time.Millisecond)

	// Retrieve the specific failed job
	reason, err := singleFailQueue.GetFailedJob("specific_fail")
	require.Nil(t, err)
	require.Contains(t, reason, "specific error for job specific_fail")

	// Try to get a non-existent failed job
	_, err = singleFailQueue.GetFailedJob("non_existent")
	require.NotNil(t, err)
	require.Contains(t, err.Error(), "not found")

	// Clean up
	err = singleFailQueue.ClearFailedJobs()
	require.Nil(t, err)
}

func Test_ClearFailedJobs(t *testing.T) {
	clearQueue := queue.New("clear_test", &queue.Options{
		Connect: &redis.Options{
			Addr:     "localhost:6379",
			Password: "",
			DB:       0,
		},
		Workers:       3,
		RetryFailures: 0,
	})

	// Clear any existing failed jobs first
	err := clearQueue.ClearFailedJobs()
	require.Nil(t, err)

	clearQueue.Process(func(job *queue.Job) {
		job.Process(func() error {
			return errors.New("error for clearing test")
		})
	})

	// Add multiple jobs that will fail
	clearQueue.BulkAddJob([]queue.AddJobOptions{
		{Id: "clear1", Data: "value 1"},
		{Id: "clear2", Data: "value 2"},
		{Id: "clear3", Data: "value 3"},
		{Id: "clear4", Data: "value 4"},
		{Id: "clear5", Data: "value 5"},
	})

	// Wait for jobs to be processed
	time.Sleep(500 * time.Millisecond)

	// Verify failed jobs exist
	failedJobs, err := clearQueue.GetFailedJobs()
	require.Nil(t, err)
	require.Equal(t, 5, len(failedJobs))

	// Clear all failed jobs
	err = clearQueue.ClearFailedJobs()
	require.Nil(t, err)

	// Verify all failed jobs are cleared
	failedJobs, err = clearQueue.GetFailedJobs()
	require.Nil(t, err)
	require.Equal(t, 0, len(failedJobs))

	// Clearing again should not cause an error
	err = clearQueue.ClearFailedJobs()
	require.Nil(t, err)
}

func Test_GetFailedJobs_RedisError(t *testing.T) {
	// Create a queue with invalid Redis connection
	invalidQueue := queue.New("redis_error_test", &queue.Options{
		Connect: &redis.Options{
			Addr:     "localhost:9999", // Invalid port
			Password: "",
			DB:       0,
		},
		Workers:       3,
		RetryFailures: 0,
	})

	// Attempt to get failed jobs should return an error
	// This tests that SCAN errors are propagated
	_, err := invalidQueue.GetFailedJobs()
	require.NotNil(t, err)
	require.Contains(t, err.Error(), "failed to scan Redis keys")
}

func Test_ClearFailedJobs_RedisError(t *testing.T) {
	// Create a queue with invalid Redis connection
	invalidQueue := queue.New("redis_clear_error_test", &queue.Options{
		Connect: &redis.Options{
			Addr:     "localhost:9999", // Invalid port
			Password: "",
			DB:       0,
		},
		Workers:       3,
		RetryFailures: 0,
	})

	// Attempt to clear failed jobs should return an error
	// This tests that SCAN errors are propagated
	err := invalidQueue.ClearFailedJobs()
	require.NotNil(t, err)
	require.Contains(t, err.Error(), "failed to scan Redis keys")
}
