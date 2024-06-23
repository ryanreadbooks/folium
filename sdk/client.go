package sdk

import (
	"context"
	"fmt"
)

type Client struct {
	isHttp bool
	isGrpc bool

	impl Impl
}

type ClientOpt func(*Client) error

func NewClient(opts ...ClientOpt) (*Client, error) {
	c := &Client{}

	for _, o := range opts {
		err := o(c)
		if err != nil {
			return nil, err
		}
	}

	// client is either http or grpc
	if c.isHttp && c.isGrpc {
		return nil, fmt.Errorf("sdk client is either http client or grpc client, can not be both")
	}

	return c, nil
}

type Impl interface {
	Next(ctx context.Context, key string, step uint32) (uint64, error)
}

func (c *Client) GetId(ctx context.Context, key string, step uint32) (uint64, error) {
	return c.impl.Next(ctx, key, step)
}
