package idgen

import (
	"context"
	"sync"
	"sync/atomic"

	"github.com/ryanreadbooks/folium/internal/segment/dao"
)

type segment struct {
	sync.Mutex
	key  string
	max  uint64
	step uint32
	cur  uint64
}

func newSegment(key string) (*segment, error) {
	sg := &segment{key: key}
	err := sg.fetchDB()
	return sg, err
}

// returns the current id and increment current id
func (s *segment) nextAndIncr() uint64 {
	var old, nuevo uint64
	for {
		old = atomic.LoadUint64(&s.cur)
		nuevo = old + 1
		if atomic.CompareAndSwapUint64(&s.cur, old, nuevo) {
			return old
		}
	}
}

// update max id
func (s *segment) update(max uint64, step uint32) {
	s.Lock()
	defer s.Unlock()

	s.step = step
	s.max = max + uint64(s.step)
	atomic.AddUint64(&s.cur, max) // we need to update the current id
}

// check if current id overflow
func (s *segment) overflow() bool {
	return atomic.LoadUint64(&s.cur) >= s.max
}

// fetch from db and update segment
func (s *segment) fetchDB() error {
	curId, step, err := dao.ConsumeKey(context.TODO(), s.key)
	if err != nil {
		return err
	}

	s.update(curId, step)
	return nil
}

// buffer holds two segments which dispense ids
type buffer struct {
	sync.RWMutex
	key string
	cur *segment // cur points to one or two
	one *segment
	two *segment
}

func newBuffer(key string) (*buffer, error) {
	one, err := newSegment(key)
	if err != nil {
		return nil, err
	}
	two, err := newSegment(key)
	if err != nil {
		return nil, err
	}
	b := &buffer{
		key: key,
		one: one,
		two: two,
		cur: one,
	}

	return b, nil
}

func (b *buffer) swap() {
	if b.cur == b.one {
		b.cur = b.two
	} else {
		b.cur = b.one
	}
}

func (b *buffer) curSegment() *segment {
	return b.cur
}

func (b *buffer) bakSegment() *segment {
	if b.cur == b.one {
		return b.two
	}
	return b.one
}

func (b *buffer) getId() (uint64, error) {
	for {
		b.RLock()
		val := b.curSegment().nextAndIncr()
		if val < b.curSegment().max {
			return val, nil
		}
		b.RUnlock()

		// val is overflow, we need to switch segment and get the next id again
		b.Lock()
		b.swap()
		b.Unlock()
	}
}
