package internal

import "context"

type Dispenser interface {
	Next(ctx context.Context, key string) (uint64, error)
}
