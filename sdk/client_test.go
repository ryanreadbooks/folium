package sdk

import (
	"context"
	"testing"
)

var (
	ctx = context.Background()
)

func TestClient_HttpClient(t *testing.T) {
	cli, _ := NewClient(WithHttp("localhost:9527"))
	id, err := cli.GetId(ctx, "test-biz", 0)
	if err != nil {
		t.Logf("err = %v\n", err)
		return
	}

	t.Logf("id = %d\n", id)
}

func TestClient_GrpcClient(t *testing.T) {
	cli, err := New(WithGrpcOpt("localhost:9528"))
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	id, err := cli.GetId(ctx, "test-biz", 0)
	if err != nil {
		t.Logf("err = %v\n", err)
		return
	}

	t.Logf("id = %d\n", id)
}

func TestClient_DowngradeClient(t *testing.T) {
	cli, err := New(WithGrpcOpt("localhost"), WithDowngrade())
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	id, err := cli.GetId(ctx, "test-biz", 0)
	if err != nil {
		t.Logf("err = %v\n", err)
		return
	}

	t.Logf("id = %d\n", id)
}