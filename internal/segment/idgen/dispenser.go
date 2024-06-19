package idgen

import (
	"context"
	"sync"
	"sync/atomic"
)

var (
	inited atomic.Bool
	rwMu   sync.RWMutex
)

func alreadyInited() bool {
	return inited.Load()
}

// init idgen
func Init() {

	inited.Store(true)
}

// GetNext returns the next id for key
func GetNext(ctx context.Context, key string) (uint64, error) {
	return getNext(ctx, key)
}

func getNext(ctx context.Context, key string) (uint64, error) {

	return 0, nil
}
