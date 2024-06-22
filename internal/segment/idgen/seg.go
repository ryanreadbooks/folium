package idgen

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"

	"github.com/ryanreadbooks/folium/internal/segment/dao"
)

type segment struct {
	sync.Mutex

	name string
	key  string
	cur  uint64
	max  uint64 // can not reach max
}

func (s *segment) String() string {
	if s == nil {
		return ""
	}
	return fmt.Sprintf("name: %s, cur: %d, max: %d", s.name, s.cur, s.max)
}

func newSegment(key string) *segment {
	return &segment{
		key: key,
	}
}

// returns the current id and increment current id
// make sure this is concurrency-safe from the outside
func (s *segment) nextAndIncr() uint64 {
	old := s.cur
	s.cur += 1
	return old
}

// update max id
func (s *segment) update(newCur, newMax uint64) {
	// make sure this is concurrency-safe from the outside
	s.cur = newCur
	s.max = newMax
}

// check if current id overflow
func (s *segment) overflow() bool {
	// that cur is equals to max is considered as overflow as well
	return atomic.LoadUint64(&s.cur) >= atomic.LoadUint64(&s.max)
}

func (s *segment) hitMark(watermark float64) bool {
	curWaterMark := float64(atomic.LoadUint64(&s.cur) / atomic.LoadUint64(&s.max))
	return curWaterMark >= watermark
}

// fetch from db and update segment
// newStep is a option step param which will be the new step for key
func (s *segment) fetchDB(ctx context.Context, newStep uint32) error {
	res, err := dao.TakeIdForKey(ctx, s.key, newStep)
	if err != nil {
		return err
	}

	s.update(res.Begin, res.End)
	return nil
}

func (s *segment) getCur() uint64 {
	return atomic.LoadUint64(&s.cur)
}

func (s *segment) getMax() uint64 {
	return atomic.LoadUint64(&s.max)
}
