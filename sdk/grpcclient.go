package sdk

import (
	"context"
	"fmt"

	apiv1 "github.com/ryanreadbooks/folium/api/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type grpcClient struct {
	cli apiv1.FoliumServiceClient
}

func WithGrpc(addr string) ClientOpt {
	return func(c *Client) error {
		c.isGrpc = true
		cc, err := grpc.NewClient(addr)
		if err != nil {
			return err
		}
		c.impl = &grpcClient{
			cli: apiv1.NewFoliumServiceClient(cc),
		}

		return nil
	}
}

func (c *grpcClient) Next(ctx context.Context, key string, step uint32) (uint64, error) {
	req := &apiv1.NextRequest{
		Key:  key,
		Step: step,
	}

	resp, err := c.cli.Next(ctx, req)
	if err != nil {
		grpcerr, ok := status.FromError(err)
		if ok {
			var baseErr error
			switch grpcerr.Code() {
			case codes.InvalidArgument:
				baseErr = ErrWrongRequestFormat
			case codes.Internal:
				baseErr = ErrFolium
			default:
				baseErr = ErrGetIdFailed
			}

			return 0, fmt.Errorf("%v: %v", baseErr, grpcerr.Message())
		}
		return 0, err
	}

	return resp.Id, nil
}
