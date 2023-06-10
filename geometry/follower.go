package geometry

import (
	"time"

	"github.com/go-co-op/gocron"
)

type FollowUpdate struct {
	Timestamp time.Time
	Price     float64
}

type Follower struct {
	tickerC chan FollowUpdate

	cron     *gocron.Scheduler
	interval time.Duration
	plot     Plot
}

func NewPlotFollower(plot Plot, itv time.Duration) *Follower {
	cron := gocron.NewScheduler(time.UTC)
	cron.StartAsync()

	return &Follower{
		cron:     cron,
		plot:     plot,
		interval: itv,
		tickerC:  make(chan FollowUpdate),
	}
}

func (pf *Follower) Start() error {
	_, err := pf.cron.Every(pf.interval).
		StartImmediately().
		Do(func() {
			t := time.Now()
			pf.tickerC <- FollowUpdate{
				Timestamp: t,
				Price:     pf.PriceAt(t),
			}
		})
	return err
}

func (pf *Follower) TickerC() chan FollowUpdate {
	return pf.tickerC
}

func (pf *Follower) PriceAt(t time.Time) float64 {
	return pf.plot.At(t)
}

func (pf *Follower) Stop() {
	pf.cron.Stop()
	close(pf.tickerC)
}
