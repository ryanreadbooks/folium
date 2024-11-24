package sdk

import (
	"context"
	"fmt"
)

type IClient interface {
	GetId(ctx context.Context, key string, step uint32) (uint64, error)
	Ping(ctx context.Context) error
}

type Client struct {
	isHttp bool
	isGrpc bool

	impl Impl
}

var _ IClient = (*Client)(nil)

type opt struct {
	http      string
	grpc      string
	downgrade bool
}

type Option func(o *opt)

func WithHttpOpt(addr string) Option {
	return func(o *opt) {
		o.http = addr
	}
}

func WithGrpcOpt(addr string) Option {
	return func(o *opt) {
		o.grpc = addr
	}
}

func WithDowngrade() Option {
	return func(o *opt) {
		o.downgrade = true
	}
}

func New(opts ...Option) (IClient, error) {
	opt := &opt{}
	for _, o := range opts {
		o(opt)
	}

	clientOpts := make([]ClientOpt, 0)
	if opt.http != "" {
		clientOpts = append(clientOpts, WithHttp(opt.http))
	}
	if opt.grpc != "" {
		clientOpts = append(clientOpts, WithGrpc(opt.grpc))
	}

	c, err := NewClient(clientOpts...)
	if err != nil {
		if opt.downgrade {
			return &downGradedClient{}, nil
		}
		return nil, err
	}

	return c, nil
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
	Ping(ctx context.Context) error
}

func (c *Client) GetId(ctx context.Context, key string, step uint32) (uint64, error) {
	return c.impl.Next(ctx, key, step)
}

func (c *Client) Ping(ctx context.Context) error {
	return c.impl.Ping(ctx)
}
