// Metrics - measure average time and rate.
package metrics

import (
	"strconv"
	"sync"
	"time"
)

// Timer satisifies the expvar.Var interface.  Tracks the average time.
// Avg is a buffered channel of length 1 that is updated with the average
// when it is calculated.  It is optional to listen to the channel for metrics.
// If the channel is not empty the channel update is
// skipped.
type Timer struct {
	count   int
	time    float64
	average float64
	mu      sync.RWMutex
	Avg     chan float64
}

// Init the Timer.  The average time(s) is calculated every period.
func (t *Timer) Init(period time.Duration) {
	t.Avg = make(chan float64, 1)
	go t.avg(period)
}

// avg sets the average time every period.
func (t *Timer) avg(period time.Duration) {
	for {
		time.Sleep(period)
		t.mu.Lock()
		if t.count == 0 {
			t.average = 0
		} else {
			t.average = t.time / float64(t.count)
		}
		t.time = 0
		t.count = 0
		if len(t.Avg) < cap(t.Avg) {
			t.Avg <- t.average
		}
		t.mu.Unlock()
	}
}

// Inc increments the timer with the duration from start to the call to Inc.
func (t *Timer) Inc(start time.Time) {
	t.mu.Lock()
	defer t.mu.Unlock()
	dt := time.Since(start).Seconds()
	t.count++
	t.time += dt
}

func (t *Timer) String() string {
	t.mu.Lock()
	defer t.mu.Unlock()
	return strconv.FormatFloat(t.average, 'g', -1, 64)
}

// Rate satisfies the expvar.Var interface.  Tracks the average rate.
// Avg is a buffered channel of length 1 that is updated with the average
// when it is calculated.  It is optional to listen to the channel for metrics.
// If the channel is not empty the channel update is
// skipped.
type Rate struct {
	count   int
	average float64
	mu      sync.RWMutex
	Avg     chan float64
}

// Init initialises Rate.  The average rate per interval is calculated every period.
func (r *Rate) Init(interval time.Duration, period time.Duration) {
	r.Avg = make(chan float64, 1)
	go r.avg(interval, period)
}

// avg loops for ever and sets the average count per interval every period
// Returns immediately if called with interval == 0
func (r *Rate) avg(interval time.Duration, period time.Duration) {
	if interval.Seconds() == 0 {
		return
	}

	for {
		time.Sleep(period)
		r.mu.Lock()
		r.average = float64(r.count) / (period.Seconds() / interval.Seconds())
		r.count = 0
		if len(r.Avg) < cap(r.Avg) {
			r.Avg <- r.average
		}
		r.mu.Unlock()
	}
}

// Inc increments the counter by 1.
func (r *Rate) Inc() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.count++
}

func (r *Rate) String() string {
	r.mu.Lock()
	defer r.mu.Unlock()
	return strconv.FormatFloat(r.average, 'g', -1, 64)
}
