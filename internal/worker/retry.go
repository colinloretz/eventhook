package worker

import "time"

var backoffSchedule = []time.Duration{
	30 * time.Second,
	5 * time.Minute,
	30 * time.Minute,
	2 * time.Hour,
	5 * time.Hour,
}

// NextBackoff returns the delay before the next attempt, or -1 if no more retries.
func NextBackoff(attemptCount int) time.Duration {
	if attemptCount >= len(backoffSchedule) {
		return -1
	}
	return backoffSchedule[attemptCount]
}
