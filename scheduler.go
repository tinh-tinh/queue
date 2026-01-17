package queue

import (
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// startScheduler starts the background scheduler loop that checks for scheduled jobs.
// It polls Redis at the specified interval to find jobs ready to run.
// If a scheduler is already running, this function returns early to prevent goroutine leaks.
func (q *Queue) startScheduler(interval time.Duration) {
	if interval == 0 {
		interval = 5 * time.Second // Default polling interval
	}

	// Prevent starting scheduler if already running
	if q.schedulerRunning {
		return
	}

	// Create new ticker and done channel
	q.schedulerTicker = time.NewTicker(interval)
	q.schedulerDone = make(chan struct{})
	q.schedulerRunning = true

	go func() {
		ticker := q.schedulerTicker
		doneChan := q.schedulerDone
		for {
			select {
			case <-ticker.C:
				q.processScheduledJobs()
			case <-doneChan:
				return
			}
		}
	}()

	q.formatLog(LoggerInfo, "Scheduler started with %v interval", interval)
}

// stopScheduler stops the scheduler gracefully.
func (q *Queue) stopScheduler() {
	if !q.schedulerRunning {
		return
	}

	if q.schedulerDone != nil {
		close(q.schedulerDone)
		// Small delay to allow goroutine to exit
		time.Sleep(10 * time.Millisecond)
	}
	if q.schedulerTicker != nil {
		q.schedulerTicker.Stop()
		q.schedulerTicker = nil
	}
	q.schedulerDone = nil
	q.schedulerRunning = false
	q.formatLog(LoggerInfo, "Scheduler stopped")
}

// ScheduleJob adds a job to the scheduled set with the given run time.
// The job will be executed when the current time reaches or exceeds runAt.
func (q *Queue) ScheduleJob(jobId string, runAt time.Time) error {
	score := float64(runAt.Unix())
	_, err := q.client.ZAdd(q.ctx, q.schedulerKey, redis.Z{
		Score:  score,
		Member: jobId,
	}).Result()
	if err != nil {
		return fmt.Errorf("failed to schedule job: %w", err)
	}
	q.formatLog(LoggerInfo, "Scheduled job %s to run at %s", jobId, runAt.Format(time.RFC3339))
	return nil
}

// GetScheduledJobs retrieves all scheduled jobs with their scheduled times.
func (q *Queue) GetScheduledJobs() ([]ScheduledJobInfo, error) {
	// Get all jobs with scores
	results, err := q.client.ZRangeWithScores(q.ctx, q.schedulerKey, 0, -1).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get scheduled jobs: %w", err)
	}

	scheduledJobs := make([]ScheduledJobInfo, 0, len(results))
	for _, z := range results {
		jobId, ok := z.Member.(string)
		if !ok {
			continue
		}
		scheduledJobs = append(scheduledJobs, ScheduledJobInfo{
			JobId:     jobId,
			RunAt:     time.Unix(int64(z.Score), 0),
			Timestamp: int64(z.Score),
		})
	}

	return scheduledJobs, nil
}

// RemoveScheduledJob removes a job from the scheduled set.
func (q *Queue) RemoveScheduledJob(jobId string) error {
	_, err := q.client.ZRem(q.ctx, q.schedulerKey, jobId).Result()
	if err != nil {
		return fmt.Errorf("failed to remove scheduled job: %w", err)
	}
	q.formatLog(LoggerInfo, "Removed scheduled job %s", jobId)
	return nil
}

// processScheduledJobs checks for jobs ready to run and moves them to the waiting list.
// This method is called periodically by the scheduler loop.
func (q *Queue) processScheduledJobs() {
	now := float64(time.Now().Unix())

	// Find all jobs with score <= current timestamp
	results, err := q.client.ZRangeByScoreWithScores(q.ctx, q.schedulerKey, &redis.ZRangeBy{
		Min: "-inf",
		Max: fmt.Sprintf("%f", now),
	}).Result()

	if err != nil {
		q.formatLog(LoggerError, "Failed to get ready scheduled jobs: %v", err)
		return
	}

	if len(results) == 0 {
		return
	}

	// Process each ready job
	for _, z := range results {
		jobId, ok := z.Member.(string)
		if !ok {
			continue
		}

		// Atomically remove from scheduled set (only one instance will succeed)
		removed, err := q.client.ZRem(q.ctx, q.schedulerKey, jobId).Result()
		if err != nil || removed == 0 {
			// Another instance already processed this job
			continue
		}

		// Add job to the queue
		q.AddJob(AddJobOptions{
			Id:   jobId,
			Data: nil, // Scheduled jobs don't have data in this implementation
		})

		q.formatLog(LoggerInfo, "Moved scheduled job %s to waiting list", jobId)
	}
}

// ScheduledJobInfo contains information about a scheduled job.
type ScheduledJobInfo struct {
	JobId     string
	RunAt     time.Time
	Timestamp int64
}
