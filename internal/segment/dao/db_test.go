package dao_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/ryanreadbooks/folium/internal/segment/dao"

	"github.com/stretchr/testify/assert"
)

var (
	ctx = context.TODO()
)

func TestMain(m *testing.M) {
	dao.InitDB()
	m.Run()
	dao.CloseDB()
}

func clean() {
	_, err := dao.GetDB().ExecContext(ctx, fmt.Sprintf("DELETE FROM %s WHERE id>0", dao.TableName))
	if err != nil {
		println(err.Error())
	}
}

func TestCreate(t *testing.T) {
	defer clean()

	id, err := dao.Create(ctx, &dao.Alloc{
		Key:   "test_biz",
		CurId: 10000,
		Step:  100,
		Desc:  "test desc",
	})

	assert.Nil(t, err)
	assert.NotZero(t, id)
}

func TestQueryByKey(t *testing.T) {
	defer clean()

	id, err := dao.Create(ctx, &dao.Alloc{
		Key:   "test_biz",
		CurId: 10000,
		Step:  100,
		Desc:  "test desc",
	})

	assert.Nil(t, err)
	assert.NotZero(t, id)

	alloc, err := dao.QueryByKey(ctx, "test_biz")
	assert.Nil(t, err)
	assert.NotNil(t, alloc)
	assert.EqualValues(t, alloc.CurId, 10000)
	assert.EqualValues(t, alloc.Step, 100)
	assert.EqualValues(t, alloc.Desc, "test desc")
}

func TestQueryAll(t *testing.T) {
	defer clean()

	allocData := []*dao.Alloc{
		{
			Key:   "test_biz",
			CurId: 1000,
			Step:  1,
			Desc:  "test desc",
		},
		{
			Key:   "wint",
			CurId: 100,
			Step:  12,
			Desc:  "wint desc",
		},
		{
			Key:   "rqe",
			CurId: 1232121,
			Step:  4,
			Desc:  "req edes",
		},
	}

	for _, a := range allocData {
		dao.Create(ctx, a)
	}

	queries, err := dao.QueryAll(ctx)
	assert.Nil(t, err)
	assert.EqualValues(t, len(queries), len(allocData))
	for _, q := range queries {
		t.Logf("%+v\n", q)
	}
}
