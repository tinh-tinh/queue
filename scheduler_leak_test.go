package queue_test

import (
	"runtime"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/require"
	"github.com/tinh-tinh/queue/v2"
)

// Test_NoGoroutineLeakOnMultipleResume verifies that calling Resume multiple times
// does not leak goroutines by ensuring the old scheduler is stopped before starting a new one.
func Test_NoGoroutineLeakOnMultipleResume(t *testing.T) {
	q := queue.New("goroutine_leak_test", &queue.Options{
		Connect: &redis.Options{
			Addr:     "localhost:6379",
			Password: "",
			DB:       0,
		},
		Workers:          1,
		RetryFailures:    0,
		Pattern:          "@every 1s",
		ScheduleInterval: 1 * time.Second,
	})

	// Get initial goroutine count
	runtime.GC()
	time.Sleep(100 * time.Millisecond)
	initialGoroutines := runtime.NumGoroutine()

	// Pause and resume multiple times
	for i := 0; i < 10; i++ {
		q.Pause()
		time.Sleep(50 * time.Millisecond)
		q.Resume()
		time.Sleep(50 * time.Millisecond)
	}

	// Final pause to stop scheduler
	q.Pause()

	// Allow time for goroutines to clean up
	runtime.GC()
	time.Sleep(200 * time.Millisecond)

	finalGoroutines := runtime.NumGoroutine()

	// The number of goroutines should not have increased significantly
	// Allow for some variance (±2) due to runtime behavior
	goroutineDiff := finalGoroutines - initialGoroutines
	require.LessOrEqual(t, goroutineDiff, 2,
		"Goroutine leak detected: initial=%d, final=%d, diff=%d",
		initialGoroutines, finalGoroutines, goroutineDiff)
}

// Test_SchedulerRestartOnResume verifies that the scheduler properly restarts
// with the correct interval when Resume is called.
func Test_SchedulerRestartOnResume(t *testing.T) {
	q := queue.New("scheduler_restart_test", &queue.Options{
		Connect: &redis.Options{
			Addr:     "localhost:6379",
			Password: "",
			DB:       0,
		},
		Workers:          1,
		RetryFailures:    0,
		Pattern:          "@every 1s",
		ScheduleInterval: 1 * time.Second,
	})

	// Schedule a job
	runAt := time.Now().Add(2 * time.Second)
	err := q.ScheduleJob("test_job", runAt)
	require.Nil(t, err)

	// Pause and resume
	q.Pause()
	time.Sleep(500 * time.Millisecond)
	q.Resume()

	// Verify scheduler is working by checking if job gets processed
	processedJobs := make(map[string]bool)
	q.Process(func(job *queue.Job) {
		job.Process(func() error {
			processedJobs[job.Id] = true
			return nil
		})
	})

	// Wait for job to be processed
	time.Sleep(3 * time.Second)

	// Verify job was processed
	require.True(t, processedJobs["test_job"], "Job should have been processed after resume")
}

// Test_MultiplePauseCalls verifies that calling Pause multiple times doesn't panic.
func Test_MultiplePauseCalls(t *testing.T) {
	q := queue.New("multiple_pause_test", &queue.Options{
		Connect: &redis.Options{
			Addr:     "localhost:6379",
			Password: "",
			DB:       0,
		},
		Workers:       1,
		RetryFailures: 0,
		Pattern:       "@every 1s",
	})

	// Call Pause multiple times - should not panic
	q.Pause()
	q.Pause()
	q.Pause()

	// Verify no panic occurred
	require.True(t, true, "Multiple Pause calls should not panic")
}

// Test_MultipleResumeCalls verifies that calling Resume multiple times doesn't
// start duplicate scheduler goroutines.
func Test_MultipleResumeCalls(t *testing.T) {
	q := queue.New("multiple_resume_test", &queue.Options{
		Connect: &redis.Options{
			Addr:     "localhost:6379",
			Password: "",
			DB:       0,
		},
		Workers:       1,
		RetryFailures: 0,
		Pattern:       "@every 1s",
	})

	// Get initial goroutine count
	runtime.GC()
	time.Sleep(100 * time.Millisecond)
	initialGoroutines := runtime.NumGoroutine()

	// Pause once
	q.Pause()
	time.Sleep(50 * time.Millisecond)

	// Call Resume multiple times without Pause in between
	q.Resume()
	time.Sleep(50 * time.Millisecond)
	q.Resume() // Should be a no-op due to schedulerRunning flag
	time.Sleep(50 * time.Millisecond)
	q.Resume() // Should be a no-op due to schedulerRunning flag
	time.Sleep(50 * time.Millisecond)

	// Final pause to stop scheduler
	q.Pause()

	// Allow time for goroutines to clean up
	runtime.GC()
	time.Sleep(200 * time.Millisecond)

	finalGoroutines := runtime.NumGoroutine()

	// The number of goroutines should not have increased significantly
	// Only ONE scheduler goroutine should have been created despite multiple Resume calls
	goroutineDiff := finalGoroutines - initialGoroutines
	require.LessOrEqual(t, goroutineDiff, 2,
		"Goroutine leak detected from multiple Resume calls: initial=%d, final=%d, diff=%d",
		initialGoroutines, finalGoroutines, goroutineDiff)
}
