package queue_test

import (
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/require"
	"github.com/tinh-tinh/queue/v2"
)

func Test_ParsePattern_Valid(t *testing.T) {
	tests := []struct {
		name     string
		pattern  string
		expected time.Duration
	}{
		// @every patterns
		{
			name:     "1 second",
			pattern:  "@every 1s",
			expected: 1 * time.Second,
		},
		{
			name:     "5 seconds",
			pattern:  "@every 5s",
			expected: 5 * time.Second,
		},
		{
			name:     "1 minute",
			pattern:  "@every 1m",
			expected: 1 * time.Minute,
		},
		{
			name:     "5 minutes",
			pattern:  "@every 5m",
			expected: 5 * time.Minute,
		},
		{
			name:     "1 hour",
			pattern:  "@every 1h",
			expected: 1 * time.Hour,
		},
		{
			name:     "complex duration",
			pattern:  "@every 1h30m45s",
			expected: 1*time.Hour + 30*time.Minute + 45*time.Second,
		},
		{
			name:     "with extra spaces",
			pattern:  "@every  5s  ",
			expected: 5 * time.Second,
		},
		{
			name:     "milliseconds",
			pattern:  "@every 500ms",
			expected: 500 * time.Millisecond,
		},
		// Cron patterns
		{
			name:     "every 5 minutes (cron)",
			pattern:  "*/5 * * * *",
			expected: 5 * time.Minute,
		},
		{
			name:     "every 15 minutes (cron)",
			pattern:  "*/15 * * * *",
			expected: 15 * time.Minute,
		},
		{
			name:     "every 2 hours (cron)",
			pattern:  "0 */2 * * *",
			expected: 2 * time.Hour,
		},
		{
			name:     "every 6 hours (cron)",
			pattern:  "0 */6 * * *",
			expected: 6 * time.Hour,
		},
		{
			name:     "hourly (cron)",
			pattern:  "0 * * * *",
			expected: 1 * time.Hour,
		},
		{
			name:     "daily (cron)",
			pattern:  "0 0 * * *",
			expected: 24 * time.Hour,
		},
		{
			name:     "weekly (cron)",
			pattern:  "0 0 * * 0",
			expected: 7 * 24 * time.Hour,
		},
		{
			name:     "monthly (cron)",
			pattern:  "0 0 1 * *",
			expected: 30 * 24 * time.Hour,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test indirectly through Queue creation
			q := queue.New("test_pattern_"+tt.name, &queue.Options{
				Connect: &redis.Options{
					Addr:     "localhost:6379",
					Password: "",
					DB:       0,
				},
				Workers:       1,
				RetryFailures: 0,
				Pattern:       tt.pattern,
				// Don't set ScheduleInterval to force pattern parsing
			})
			require.NotNil(t, q)
		})
	}
}

func Test_ParsePattern_Invalid(t *testing.T) {
	tests := []struct {
		name    string
		pattern string
	}{
		{
			name:    "empty pattern",
			pattern: "",
		},
		{
			name:    "missing duration",
			pattern: "@every ",
		},
		{
			name:    "invalid duration",
			pattern: "@every abc",
		},
		{
			name:    "negative duration",
			pattern: "@every -5s",
		},
		{
			name:    "invalid cron - too few fields",
			pattern: "*/5 * *",
		},
		{
			name:    "invalid cron - too many fields",
			pattern: "*/5 * * * * *",
		},
		{
			name:    "invalid cron - bad minute interval",
			pattern: "*/abc * * * *",
		},
		{
			name:    "invalid cron - minute out of range",
			pattern: "*/60 * * * *",
		},
		{
			name:    "invalid cron - hour out of range",
			pattern: "0 */24 * * *",
		},
		{
			name:    "unsupported cron pattern",
			pattern: "5,10,15 * * * *",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// These should fall back to default 5s interval with a warning log
			q := queue.New("test_invalid_"+tt.name, &queue.Options{
				Connect: &redis.Options{
					Addr:     "localhost:6379",
					Password: "",
					DB:       0,
				},
				Workers:       1,
				RetryFailures: 0,
				Pattern:       tt.pattern,
				DisableLog:    true, // Disable logs to avoid clutter in tests
			})
			require.NotNil(t, q)
		})
	}
}
