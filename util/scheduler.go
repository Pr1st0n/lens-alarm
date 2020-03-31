package util

import (
	"time"
)

const (
	INTERVAL_PERIOD = 24 * time.Hour
	HOUR_TO_TICK    = 06
	MINUTE_TO_TICK  = 00
	SECOND_TO_TICK  = 00
)

type job func()

type jobTicker struct {
	t *time.Timer
}

func getNextTickDuration() time.Duration {
	now := time.Now()
	nextTick := time.Date(now.Year(), now.Month(), now.Day(), HOUR_TO_TICK, MINUTE_TO_TICK, SECOND_TO_TICK, 0, time.Local)

	if nextTick.Before(now) {
		nextTick = nextTick.Add(INTERVAL_PERIOD)
	}

	return nextTick.Sub(time.Now())
}

func newJobTicker() jobTicker {
	return jobTicker{time.NewTimer(getNextTickDuration())}
}

func ScheduleJob(jobFunc job) {
	ticker := newJobTicker()

	for {
		<-ticker.t.C
		jobFunc()
		ticker.t.Reset(getNextTickDuration())
	}
}
