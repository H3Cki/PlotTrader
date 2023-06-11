package geometry

import (
	"time"

	"github.com/go-co-op/gocron"
)

type FollowUpdate struct {
	Timestamp time.Time
	Price     float64
}

type FollowerOption func(*Follower)

type Follower struct {
	tickerC chan FollowUpdate

	startTime, stopTime time.Time

	cron     *gocron.Scheduler
	interval time.Duration
	plot     Plot
}

func NewPlotFollower(plot Plot, itv time.Duration, options ...FollowerOption) *Follower {
	cron := gocron.NewScheduler(time.UTC)
	cron.StartAsync()

	f := &Follower{
		cron:     cron,
		plot:     plot,
		interval: itv,
		tickerC:  make(chan FollowUpdate),
	}

	for _, opt := range options {
		opt(f)
	}

	return f
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

func WithStartTime(t time.Time) FollowerOption {
	return func(f *Follower) {
		f.startTime = t
	}
}

func WithStopTime(t time.Time) FollowerOption {
	return func(f *Follower) {
		f.stopTime = t
	}
}
