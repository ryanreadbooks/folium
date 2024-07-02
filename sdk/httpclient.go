package sdk

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/ryanreadbooks/folium/internal/pkg"
	"github.com/ryanreadbooks/folium/internal/segment/server"
)

type httpClient struct {
	c    *http.Client
	addr string
}

func WithHttp(addr string) ClientOpt {
	return func(c *Client) error {
		c.isHttp = true
		c.impl = &httpClient{
			c:    &http.Client{},
			addr: addr,
		}

		return nil
	}
}

func (c *httpClient) Next(ctx context.Context, key string, step uint32) (uint64, error) {
	path := fmt.Sprintf("http://%s/api/v1/next/%s", c.addr, key)
	if step != 0 {
		path = fmt.Sprintf("%s?step=%d", path, step)
	}

	resp, err := c.c.Get(path)
	if err != nil {
		// network error
		return 0, err
	}

	// resp contains the result of the request
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	var result server.Result
	err = json.Unmarshal(body, &result)
	if err != nil {
		return 0, fmt.Errorf("%v: %v: statuscode: %d", ErrResultNotRecognized, err, resp.StatusCode)
	}

	if result.Id == 0 {
		// folium error occur
		var errMsg pkg.Err
		err = json.Unmarshal([]byte(result.Msg), &errMsg)
		if err != nil {
			return 0, fmt.Errorf("%v: %v: statuscode: %d", ErrResultNotRecognized, err, resp.StatusCode)
		}
		return 0, errMsg
	}

	return result.Id, nil
}

func (c *httpClient) Ping(ctx context.Context) error {
	path := fmt.Sprintf("http://%s/api/v1/health", c.addr)
	resp, err := c.c.Get(path)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("ping err: statuscode: %d", resp.StatusCode)
	}

	return nil
}
