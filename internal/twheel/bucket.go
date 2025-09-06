package twheel

import (
	"container/list"
	"sync"
)

type bucket struct {
	mu    sync.Mutex
	lst   *list.List
	index map[uint64]*list.Element
}

func newBucket() *bucket {
	return &bucket{lst: list.New(), index: make(map[uint64]*list.Element)}
}

func (b *bucket) add(t *timer) {
	b.mu.Lock()
	defer b.mu.Unlock()

	el := b.lst.PushBack(t)

	t.elem = el
	t.bucket = b

	b.index[t.id] = el
}

func (b *bucket) removeByID(id uint64) *timer {
	b.mu.Lock()
	defer b.mu.Unlock()

	el, ok := b.index[id]
	if !ok {
		return nil
	}

	delete(b.index, id)

	val := b.lst.Remove(el)
	if val == nil {
		return nil
	}

	t := val.(*timer)
	t.bucket, t.elem = nil, nil

	return t
}

func (b *bucket) drainDue() (due []*timer, dec []*timer) {
	b.mu.Lock()
	defer b.mu.Unlock()

	for e := b.lst.Front(); e != nil; {
		n := e.Next()
		t := e.Value.(*timer)

		if t.rounds > 0 {
			t.rounds--
			dec = append(dec, t)
		} else {
			delete(b.index, t.id)
			b.lst.Remove(e)
			t.bucket, t.elem = nil, nil
			due = append(due, t)
		}
		e = n
	}

	return
}
