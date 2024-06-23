package idgen

import (
	"context"
	"log"
	"sync"
	"time"
)

const (
	watermark = 0.85
)

// buffer holds two segments which dispense ids
type buffer struct {
	sync.RWMutex

	key  string
	cur  *segment // cur points to seg1 or seg2
	seg1 *segment
	seg2 *segment

	closeCh chan struct{}
	step    uint32 // step for changing the step in db
	ctx     context.Context
	cancel  context.CancelFunc
}

func newBuffer(ctx context.Context, key string, step uint32) (*buffer, error) {
	seg1 := newSegment(key)
	err := seg1.fetchDB(ctx, step) // need to be synced with db
	if err != nil {
		return nil, err
	}
	seg1.name = "seg1"
	log.Printf("seg1 loaded with %+v\n", seg1)
	seg2 := newSegment(key)
	seg2.name = "seg2"

	cctx, ccancel := context.WithCancel(context.Background())
	b := &buffer{
		key:     key,
		seg1:    seg1,
		seg2:    seg2, // we do not fetchDB in the first place
		cur:     seg1,
		closeCh: make(chan struct{}),
		step:    step,
		ctx:     cctx,
		cancel:  ccancel,
	}

	// go b.worker()

	return b, nil
}

// swap without lock
func (b *buffer) swap(ctx context.Context) error {
	if b.cur == b.seg1 {
		// make sure we are swapping into maxinum segment
		if b.seg1.max > b.seg2.max {
			err := b.seg2.fetchDB(ctx, b.step)
			if err != nil {
				log.Printf("buffer swap to seg2 fetchDB err: %v\n", err)
				return err
			}
			log.Printf("buffer swap to seg2 fetchDB: %+v\n", b.seg2)
		}
		b.cur = b.seg2
	} else {
		if b.seg1.max < b.seg2.max {
			err := b.seg1.fetchDB(ctx, b.step)
			if err != nil {
				log.Printf("buffer swap to seg1 fetchDB err: %v\n", err)
				return err
			}
			log.Printf("buffer swap to seg1 fetchDB: %+v\n", b.seg1)
		}
		b.cur = b.seg1
	}

	return nil
}

func (b *buffer) curSeg() *segment {
	return b.cur
}

func (b *buffer) bakSeg() *segment {
	if b.cur == b.seg1 {
		return b.seg2
	}
	return b.seg1
}

func (b *buffer) getId(ctx context.Context) (uint64, error) {
	for {
		b.Lock()
		curSeg := b.curSeg()
		val := curSeg.nextAndIncr()
		if val < curSeg.max {
			b.Unlock()
			return val, nil
		}

		// val is overflow, we need to switch segment and get the next id again
		var err error = b.swap(ctx)
		if err != nil {
			log.Printf("buffer getId swap err: %v\n", err)
			b.Unlock()
			return 0, err
		}
		b.Unlock()
	}
}

// preload will check and do swapping stuff after get Id
// just to prevent worker is not working properly
func (b *buffer) preload() {
	defer func() {
		if err := recover(); err != nil {
			log.Printf("buffer postGetId panic: %v\n", err)
		}
	}()

	b.Lock()
	defer b.Unlock()

	// we hit watermark
	if b.curSeg().hitMark(watermark) {
		// load the other segment
		err := b.bakSeg().fetchDB(b.ctx, b.step)
		if err != nil {
			log.Printf("buffer postGetId fetchDB err: %v\n", err)
			return
		}
		log.Printf("buffer loaded segment updated: %+v\n", b.bakSeg())
	}
}

// worker the situation of two segments
func (b *buffer) worker() {
	defer func() {
		if err := recover(); err != nil {
			log.Printf("buffer monitor panic: %v\n", err)
			go b.worker()
		}
	}()

	ticker := time.NewTicker(time.Millisecond * 500)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			b.preload()
		case <-b.closeCh:
			log.Println("buffer worker exited")
			b.cancel()
			return
		}
	}
}

func (b *buffer) close() {
	b.closeCh <- struct{}{}
}
