package dao

import (
	"context"
	"fmt"
	"math/rand"
	"sync"
	"testing"

	"github.com/ryanreadbooks/folium/internal/pkg/misc"
	"github.com/stretchr/testify/assert"
)

var (
	ctx = context.TODO()
)

func TestMain(m *testing.M) {
	InitDB()
	m.Run()
	CloseDB()
}

func clean() {
	_, err := GetDB().ExecContext(ctx, fmt.Sprintf("DELETE FROM %s WHERE id>0", TableName))
	if err != nil {
		println(err.Error())
	}
}

func TestUpdate(t *testing.T) {
	defer clean()

	err := CreateUpdate(ctx, &Alloc{
		Key:   "test_biz",
		CurId: 10000,
		Step:  100,
	})

	assert.Nil(t, err)
}

func TestQueryByKey(t *testing.T) {
	defer clean()
	err := CreateUpdate(ctx, &Alloc{
		Key:   "test_biz",
		CurId: 10000,
		Step:  100,
	})

	assert.Nil(t, err)

	alloc, err := QueryByKey(ctx, "test_biz")
	assert.Nil(t, err)
	assert.NotNil(t, alloc)
	assert.EqualValues(t, alloc.CurId, 10000)
	assert.EqualValues(t, alloc.Step, 100)

	_, err = QueryByKey(ctx, "not-found")
	t.Log(err)
}

func TestQueryAll(t *testing.T) {
	defer clean()

	allocData := []*Alloc{
		{
			Key:   "test_biz",
			CurId: 1000,
			Step:  1,
		},
		{
			Key:   "wint",
			CurId: 100,
			Step:  12,
		},
		{
			Key:   "rqe",
			CurId: 1232121,
			Step:  4,
		},
	}

	for _, a := range allocData {
		CreateUpdate(ctx, a)
	}

	queries, err := QueryAll(ctx)
	assert.Nil(t, err)
	assert.EqualValues(t, len(queries), len(allocData))
	for _, q := range queries {
		t.Logf("%+v\n", q)
	}
}

func TestConsumeKey(t *testing.T) {
	defer clean()

	res, err := TakeIdForKey(ctx, "test-biz", 0)
	assert.Nil(t, err)
	assert.EqualValues(t, res.Begin, defaultCurId)
	assert.EqualValues(t, res.Step, defaultStep)

	var (
		keys = []string{"biz1", "biz2", "biz3"}
	)

	num := 100
	curIdMap := make(map[string][]uint64)
	wg := sync.WaitGroup{}
	lk := sync.Mutex{}
	for i := 0; i < num; i++ {
		wg.Add(1)
		t.Run(fmt.Sprintf("case-%d", i), func(t *testing.T) {
			defer wg.Done()

			key := keys[rand.Intn(len(keys))]
			res, err := TakeIdForKey(ctx, key, 0)
			assert.Nil(t, err)
			lk.Lock()
			curIdMap[key] = append(curIdMap[key], res.Begin)
			lk.Unlock()
		})
	}

	wg.Wait()

	for k, v := range curIdMap {
		// remote duplicate in v
		assert.EqualValues(t, misc.HasDupElems(v), false)
		t.Logf("k = %s, v = %v\n", k, v)
	}

}
