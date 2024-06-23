package idgen

import (
	"context"
	"sync"
	"sync/atomic"

	"github.com/ryanreadbooks/folium/internal/pkg"
	"github.com/ryanreadbooks/folium/internal/segment/dao"
	"google.golang.org/grpc/codes"
)

const (
	maxStepAllowed = 100000
)

var (
	rwMu   sync.Mutex
	closed atomic.Bool

	bufs sync.Map
)

var (
	ErrClosed = pkg.NewErr(int(codes.Unavailable), "segment idgen dispenser is closed")
)

// init idgen
func Init() {
	dao.InitDB()
	closed.Store(false)
}

type GetOption struct {
	Step uint32
}

type Option func(*GetOption)

func WithStep(step uint32) Option {
	return func(o *GetOption) {
		o.Step = step
		if o.Step > maxStepAllowed {
			o.Step = maxStepAllowed
		}
	}
}

// GetNext returns the next id for key
func GetNext(ctx context.Context, key string, opt ...Option) (uint64, error) {
	if closed.Load() {
		return 0, ErrClosed
	}

	if len(key) == 0 {
		return 0, pkg.ErrInvalidArgs.Message("key is empty")
	}

	gOpt := &GetOption{}
	for _, o := range opt {
		o(gOpt)
	}

	var (
		buf *buffer
		ok  bool
		err error
	)

	val, ok := bufs.Load(key)
	if !ok {
		// buf is new here, we need to create it now
		buf, err = newBuffer(ctx, key, gOpt.Step)
		if err != nil {
			return 0, pkg.ErrInternal
		}
		bufs.Store(key, buf)
	} else {
		buf, ok = val.(*buffer)
		if !ok {
			return 0, pkg.ErrInternal.Message("segment buffer type mismatch")
		}
	}

	id, err := buf.getId(ctx)
	if err != nil {
		return 0, pkg.ErrInternal
	}

	return id, nil
}

func Close() {
	closed.Store(true)
	dao.CloseDB()
}
