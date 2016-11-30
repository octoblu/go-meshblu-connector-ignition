package interval

import "time"

// OnInterval is fired once every delay passes. If OnInterval takes
// longer than the delay, the next execution will wait until the
// current one has been completed.
type OnInterval func()

// Interval allows a running interval to be stopped
type Interval interface {

	// Clear clears the delay. If the OnInterval `fp` is currently being
	// called, it will allow it to finish, but it will not be called again.
	// May safely be called multiple times
	Clear()
}

type tickerInterval struct {
	delay   time.Duration
	fp      OnInterval
	cleared bool
}

// SetInterval calls the OnInterval function pointer `fp` every `delay`
func SetInterval(delay time.Duration, fp OnInterval) Interval {
	interval := &tickerInterval{
		delay:   delay,
		fp:      fp,
		cleared: false,
	}
	go interval.run()
	return interval
}

func (interval *tickerInterval) Clear() {
	interval.cleared = true
}

func (interval *tickerInterval) run() {
	ticker := time.NewTicker(interval.delay)

	for range ticker.C {
		if interval.cleared {
			ticker.Stop()
			return
		}
		interval.fp()
	}
}
