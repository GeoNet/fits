// Metrics - measure time and rate.
package metrics

import (
	"expvar"
	"log"
	"time"
)

type Timer struct {
	count  int
	time   float64
	Period time.Duration // avg is calculated at period.
	V      *expvar.Float
}

// Avg sets V to the average time every period.
// Run as a goroutine  once the Timer is configured.
//
//    dbTime := metrics.Timer{Period: 30 * time.Second, V: expvar.NewFloat("averageDBResponseTime")}
//    go dbTime.Avg()
//
func (t *Timer) Avg() {
	for {
		time.Sleep(t.Period)
		// Local copy of count so there is no
		// risk of div by 0 errors.
		n := t.count
		if n == 0 {
			t.V.Set(0)
		} else {
			t.V.Set(t.time / float64(n))
		}
		t.time = 0
		t.count = 0
	}
}

// Track logs message and time since start.  Increments the timer counters.
func (t *Timer) Track(start time.Time, message string) {
	dt := time.Since(start).Seconds()
	log.Printf("%s took %fs", message, dt)
	t.count++
	t.time += dt
}

type Rate struct {
	count    int
	Period   time.Duration // avg is calculated at interval.
	Interval time.Duration // V is set to count per Interval every Period.
	V        *expvar.Float
}

// avg loops for ever and sets V to the average count per Interval every Period
func (c *Rate) Avg() {
	for {
		time.Sleep(c.Period)
		if c.Period.Seconds() == 0 {
			c.V.Set(0)
		} else {
			c.V.Set(float64(c.count) / (c.Period.Seconds() / c.Interval.Seconds()))
		}
		c.count = 0
	}
}

// Inc increments the counter by 1.
func (c *Rate) Inc() {
	c.count++
}
