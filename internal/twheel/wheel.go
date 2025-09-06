package twheel

import (
	"container/list"
	"context"
	"sync"
	"sync/atomic"
	"time"
)

type Task func()

type timer struct {
	id       uint64
	deadline time.Time // UTC
	rounds   int
	task     Task

	bucket *bucket
	elem   *list.Element
}

type addReq struct {
	id       uint64
	deadline time.Time
	task     Task
}

type cancelReq struct {
	id     uint64
	result chan bool
}

type Wheel struct {
	tick  time.Duration
	slots int
	wheel []*bucket

	ticker *time.Ticker
	cur    int

	addCh    chan addReq
	cancelCh chan cancelReq
	stopCh   chan struct{}
	wg       sync.WaitGroup

	idGen  atomic.Uint64
	timers sync.Map // id -> *timer
}

func New(tick time.Duration, slots int) *Wheel {
	if tick <= 0 || slots < 2 {
		panic("invalid wheel config")
	}

	w := &Wheel{
		tick:     tick,
		slots:    slots,
		wheel:    make([]*bucket, slots),
		addCh:    make(chan addReq, 2048),
		cancelCh: make(chan cancelReq, 2048),
		stopCh:   make(chan struct{}),
	}

	for i := 0; i < slots; i++ {
		w.wheel[i] = newBucket()
	}

	return w
}

func (w *Wheel) Start() {
	if w.ticker != nil {
		return
	}

	w.ticker = time.NewTicker(w.tick)
	w.wg.Add(1)

	go w.loop()
}
func (w *Wheel) Stop(ctx context.Context) error {
	if w.ticker == nil {
		return nil
	}

	close(w.stopCh)
	done := make(chan struct{})

	go func() { w.wg.Wait(); close(done) }()

	select {
	case <-done:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func ceilDiv(d, base time.Duration) int {
	if d <= 0 {
		return 0
	}

	n := int(d / base)

	if d%base != 0 {
		n++
	}

	return n
}

func (w *Wheel) loop() {
	defer w.wg.Done()

	for {
		select {
		case <-w.ticker.C:
			b := w.wheel[w.cur]
			due, _ := b.drainDue()
			for _, t := range due {
				w.timers.Delete(t.id)
				go func(t *timer) {
					defer func() { _ = recover() }()
					t.task()
				}(t)
			}
			w.cur = (w.cur + 1) % w.slots
		consume:
			for {
				select {
				case a := <-w.addCh:
					w.place(a)
				case c := <-w.cancelCh:
					ok := w.doCancel(c.id)
					if c.result != nil {
						c.result <- ok
					}
				default:
					break consume
				}
			}
		case a := <-w.addCh:
			w.place(a)
		case c := <-w.cancelCh:
			ok := w.doCancel(c.id)
			if c.result != nil {
				c.result <- ok
			}
		case <-w.stopCh:
			w.ticker.Stop()
			return
		}
	}
}

func (w *Wheel) place(a addReq) {
	now := time.Now().UTC()
	delay := a.deadline.Sub(now)

	ticksLater := max(ceilDiv(delay, w.tick), 0)

	slot := (w.cur + (ticksLater % w.slots)) % w.slots
	rounds := ticksLater / w.slots

	t := &timer{id: a.id, deadline: a.deadline, rounds: rounds, task: a.task}

	w.timers.Store(t.id, t)
	w.wheel[slot].add(t)
}

func (w *Wheel) doCancel(id uint64) bool {
	if v, ok := w.timers.Load(id); ok {
		t := v.(*timer)
		if t.bucket != nil {
			_ = t.bucket.removeByID(id)
		}

		w.timers.Delete(id)

		return true
	}

	return false
}

func (w *Wheel) AfterFunc(d time.Duration, f Task) uint64 {
	id := w.idGen.Add(1)
	w.addCh <- addReq{id: id, deadline: time.Now().UTC().Add(d), task: f}

	return id
}

func (w *Wheel) At(deadline time.Time, f Task) uint64 {
	id := w.idGen.Add(1)
	w.addCh <- addReq{id: id, deadline: deadline.UTC(), task: f}

	return id
}

func (w *Wheel) Cancel(id uint64) bool {
	res := make(chan bool, 1)
	w.cancelCh <- cancelReq{id: id, result: res}

	return <-res
}
