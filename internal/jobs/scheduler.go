package jobs

import (
	"context"
	"log"
	"time"

	"github.com/yplog/ticktockbox/internal/rmq"
	"github.com/yplog/ticktockbox/internal/twheel"
)

type Scheduler struct {
	Repo *Repo
	Pub  *rmq.Publisher
	Wh   *twheel.Wheel
}

type DueEvent struct {
	ID       int64     `json:"id"`
	Title    string    `json:"title"`
	RunAtUTC time.Time `json:"run_at_utc"`
	DueAtUTC time.Time `json:"due_at_utc"`
	TZ       string    `json:"tz"`
}

func NewScheduler(repo *Repo, pub *rmq.Publisher, wh *twheel.Wheel) *Scheduler {
	return &Scheduler{Repo: repo, Pub: pub, Wh: wh}
}

func (s *Scheduler) Warmup(ctx context.Context) error {
	since := time.Now().UTC().Add(-10 * time.Minute)
	pending, err := s.Repo.LoadPendingSince(ctx, since)
	if err != nil {
		return err
	}
	for _, j := range pending {
		j := j
		s.scheduleJob(j)
	}
	return nil
}

func (s *Scheduler) ScheduleNew(j Job) {
	s.scheduleJob(j)
}

func (s *Scheduler) scheduleJob(j Job) {
	deadline := j.DueAtUTC
	if deadline.Before(time.Now().UTC()) {
		deadline = time.Now().UTC()
	}

	s.Wh.At(deadline, func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		ev := DueEvent{ID: j.ID, Title: j.Title, RunAtUTC: j.RunAtUTC, DueAtUTC: j.DueAtUTC, TZ: j.TZ}
		if err := s.Pub.PublishJSON(ctx, ev, keyFor(j.ID)); err != nil {
			log.Printf("publish failed job=%d err=%v", j.ID, err)
			return
		}
		if err := s.Repo.MarkEnqueued(context.Background(), j.ID); err != nil {
			log.Printf("mark enqueued failed job=%d err=%v", j.ID, err)
		}
	})
}

func keyFor(id int64) string { return "job-" + time.Now().UTC().Format("20060102150405") }
