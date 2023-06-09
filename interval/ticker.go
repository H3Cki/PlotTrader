package interval

import (
	"time"

	"github.com/go-co-op/gocron"
)

type Tick struct {
	Timestamp time.Time
}

type Ticker struct {
	cron *gocron.Scheduler
}

// NewTicker Creates a notifier and starts it's asynchronous loop
func NewTicker() *Ticker {
	c := &Ticker{
		cron: gocron.NewScheduler(time.UTC),
	}

	c.cron.StartAsync()

	return c
}

// Add adds a new cron job for given interval starting at given time and returns a channel
func (c *Ticker) Add(start time.Time, every time.Duration) (notifyC chan Tick, stop func(), err error) {
	notifC := make(chan Tick)

	job, err := c.cron.Every(every).
		StartAt(start).
		Do(func() {
			notifC <- Tick{
				Timestamp: time.Now().UTC(),
			}
		})

	if err != nil {
		return nil, nil, err
	}

	stop = func() {
		c.cron.RemoveByReference(job)
		close(notifC)
	}

	return notifC, stop, nil
}
