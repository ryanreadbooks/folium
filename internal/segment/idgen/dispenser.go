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

// GetNext returns the next id for key
func GetNext(ctx context.Context, key string) (uint64, error) {
	if closed.Load() {
		return 0, ErrClosed
	}

	if len(key) == 0 {
		return 0, pkg.ErrInvalidArgs.Message("key is empty")
	}

	var (
		buf *buffer
		ok  bool
		err error
	)

	val, ok := bufs.Load(key)
	if !ok {
		// buf is new here, we need to create it now
		buf, err = newBuffer(ctx, key)
		if err != nil {
			return 0, pkg.ErrInternal.Message(err.Error())
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
		return 0, pkg.ErrInternal.Message(err.Error())
	}

	return id, nil
}

func GetNextWithStep(ctx context.Context, key string, step uint32) (uint64, error) {
	if closed.Load() {
		return 0, ErrClosed
	}

	if len(key) == 0 {
		return 0, pkg.ErrInvalidArgs.Message("key is empty")
	}

	if step > maxStepAllowed {
		return 0, pkg.ErrInvalidArgs.Message("step is too large")
	}

	// make sure step becomes effective

	return 0, nil
}

func Close() {
	closed.Store(true)
	dao.CloseDB()
}
