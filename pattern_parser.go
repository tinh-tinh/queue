package queue

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// parsePattern parses a cron-like pattern and returns the polling interval.
// Supports two formats:
// 1. @every <duration> format (e.g., "@every 1s", "@every 5m")
// 2. Cron expressions (e.g., "*/5 * * * *", "0 */2 * * *")
// Returns an error if the pattern is invalid or unsupported.
func parsePattern(pattern string) (time.Duration, error) {
	if pattern == "" {
		return 0, fmt.Errorf("pattern cannot be empty")
	}

	// Trim whitespace
	pattern = strings.TrimSpace(pattern)

	// Check for @every prefix
	if strings.HasPrefix(pattern, "@every ") {
		return parseEveryPattern(pattern)
	}

	// Try to parse as cron expression
	return parseCronPattern(pattern)
}

// parseEveryPattern parses @every <duration> format patterns.
func parseEveryPattern(pattern string) (time.Duration, error) {
	// Extract the duration part after "@every "
	durationStr := strings.TrimSpace(strings.TrimPrefix(pattern, "@every "))
	if durationStr == "" {
		return 0, fmt.Errorf("missing duration in pattern: %s", pattern)
	}

	// Parse the duration using time.ParseDuration
	duration, err := time.ParseDuration(durationStr)
	if err != nil {
		return 0, fmt.Errorf("invalid duration '%s': %w", durationStr, err)
	}

	// Validate that duration is positive
	if duration <= 0 {
		return 0, fmt.Errorf("duration must be positive, got: %s", duration)
	}

	return duration, nil
}

// parseCronPattern parses cron expressions and calculates the polling interval.
// Supports standard 5-field cron format: minute hour day month weekday
// Examples:
//   - "*/5 * * * *" → every 5 minutes
//   - "0 * * * *" → every hour
//   - "0 0 * * *" → every day (24 hours)
//   - "0 0 * * 0" → every week (7 days)
func parseCronPattern(pattern string) (time.Duration, error) {
	fields := strings.Fields(pattern)
	if len(fields) != 5 {
		return 0, fmt.Errorf("invalid cron expression: expected 5 fields, got %d in '%s'", len(fields), pattern)
	}

	minute, hour, day, month, weekday := fields[0], fields[1], fields[2], fields[3], fields[4]

	// Parse minute field for */N patterns
	if strings.HasPrefix(minute, "*/") {
		intervalStr := strings.TrimPrefix(minute, "*/")
		interval, err := strconv.Atoi(intervalStr)
		if err != nil {
			return 0, fmt.Errorf("invalid minute interval '%s': %w", intervalStr, err)
		}
		if interval <= 0 || interval > 59 {
			return 0, fmt.Errorf("minute interval must be between 1 and 59, got %d", interval)
		}
		return time.Duration(interval) * time.Minute, nil
	}

	// Parse hour field for */N patterns
	if strings.HasPrefix(hour, "*/") {
		intervalStr := strings.TrimPrefix(hour, "*/")
		interval, err := strconv.Atoi(intervalStr)
		if err != nil {
			return 0, fmt.Errorf("invalid hour interval '%s': %w", intervalStr, err)
		}
		if interval <= 0 || interval > 23 {
			return 0, fmt.Errorf("hour interval must be between 1 and 23, got %d", interval)
		}
		return time.Duration(interval) * time.Hour, nil
	}

	// Hourly: "0 * * * *" or "N * * * *"
	if hour == "*" && day == "*" && month == "*" && weekday == "*" {
		return 1 * time.Hour, nil
	}

	// Daily: "0 0 * * *" or "N N * * *"
	if day == "*" && month == "*" && weekday == "*" {
		return 24 * time.Hour, nil
	}

	// Weekly: "0 0 * * N" (specific weekday)
	if day == "*" && month == "*" && weekday != "*" {
		return 7 * 24 * time.Hour, nil
	}

	// Monthly: "0 0 N * *" (specific day of month)
	if month == "*" && weekday == "*" && day != "*" {
		return 30 * 24 * time.Hour, nil // Approximate as 30 days
	}

	return 0, fmt.Errorf("unsupported cron pattern: %s (consider using @every <duration> format)", pattern)
}
