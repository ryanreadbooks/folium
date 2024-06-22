package idgen

// test segment struct

import (
	"context"
	"fmt"
	"math/rand"
	"sync"
	"testing"
	"time"

	"github.com/ryanreadbooks/folium/internal/pkg/misc"
	"github.com/ryanreadbooks/folium/internal/segment/dao"
	"github.com/stretchr/testify/assert"
)

func TestMain(m *testing.M) {
	dao.InitDB()
	m.Run()
	dao.CloseDB()
}

var (
	ctx = context.Background()
)

func clean() {
	_, err := dao.GetDB().ExecContext(ctx, fmt.Sprintf("DELETE FROM %s WHERE id>0", dao.TableName))
	if err != nil {
		println(err.Error())
	}
}

func TestSegment_fetchDB(t *testing.T) {
	defer clean()

	seg := newSegment("biz-test")
	err := seg.fetchDB(ctx, 0)
	assert.Nil(t, err)
	assert.EqualValues(t, 1, seg.cur)
	assert.EqualValues(t, 1001, seg.max)
	t.Logf("seg = %+v\n", seg)
}

func TestSegment_nextAndIncr(t *testing.T) {
	seg := &segment{
		key: "biz-ztest",
		cur: 1,
		max: 1000,
	}

	num := 200
	var wg sync.WaitGroup
	var lock sync.Mutex
	ids := make([]uint64, 0, num*2)
	for i := 0; i < num; i++ {
		wg.Add(1)
		go func(tt *testing.T) {
			defer wg.Done()
			ms := rand.Intn(10)
			time.Sleep(time.Millisecond * time.Duration(ms+1))
			id := seg.nextAndIncr()
			lock.Lock()
			ids = append(ids, id)
			lock.Unlock()
		}(t)
	}

	wg.Wait()

	assert.EqualValues(t, misc.HasDupElems(ids), false)
	t.Logf("ids = %v\n", ids)

	wg.Add(1)
	var wg2 sync.WaitGroup
	wg2.Add(2)
	go func() {
		defer wg2.Done()
		wg.Wait()
		t.Logf("g1 = %d\n", seg.nextAndIncr())
	}()

	go func() {
		defer wg2.Done()
		wg.Wait()
		t.Logf("g2 = %d\n", seg.nextAndIncr())
	}()

	time.Sleep(time.Millisecond*10)
	wg.Done()
	wg2.Wait()
}
