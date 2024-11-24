package sdk

import "context"

type downGradedClient struct {
}

var _ IClient = (*downGradedClient)(nil)

func (c *downGradedClient) GetId(ctx context.Context, key string, step uint32) (uint64, error) {
	return 0, ErrFoliumNotConnected
}

func (c *downGradedClient) Ping(ctx context.Context) error {
	return ErrFoliumNotConnected
}
