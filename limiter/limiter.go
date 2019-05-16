package limiter

import (
	"time"
)

type Limiter struct {
	limitCounter int64
	startTime    time.Time
	rps          float64
	npr          int64
}

func New(rps float64) *Limiter {
	return &Limiter{
		limitCounter: 1,
		startTime:    time.Now(),
		rps:          rps,
		npr:          int64(1000000000 / rps),
	}
}

func (l *Limiter) Sleep() {
	if l.rps <= 0 {
		return
	}
	toSleep := l.limitCounter*l.npr - time.Now().Sub(l.startTime).Nanoseconds()
	time.Sleep(time.Duration(toSleep))
	l.limitCounter++
}

func (l *Limiter) Reset() {
	l.limitCounter = 1
	l.startTime = time.Now()
}
