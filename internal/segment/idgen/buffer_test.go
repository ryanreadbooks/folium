package idgen

import (
	"math/rand"
	"sync"
	"testing"
	"time"

	"github.com/ryanreadbooks/folium/internal/pkg/misc"
	"github.com/stretchr/testify/assert"
)

func TestBuffer_newBuffer(t *testing.T) {
	defer clean()

	_, err := newBuffer(ctx, "biz-test", 0)
	assert.Nil(t, err)
}

func TestBuffer_getId(t *testing.T) {
	defer clean()

	buf, err := newBuffer(ctx, "biz-test", 0)
	assert.Nil(t, err)
	assert.NotNil(t, buf)

	id, err := buf.getId(ctx)
	assert.Nil(t, err)
	t.Logf("getId = %d\n", id)
}

func TestBuffer_concurrentGetId(t *testing.T) {
	defer clean()

	buf, err := newBuffer(ctx, "biz-test", 0)
	assert.Nil(t, err)
	assert.NotNil(t, buf)

	var wg sync.WaitGroup
	num := 100000
	var lock sync.Mutex

	ids := make([]uint64, 0, num*2)
	for i := 0; i < num; i++ {
		wg.Add(1)
		go func(tt *testing.T) {
			defer wg.Done()
			ms := rand.Intn(10)
			time.Sleep(time.Millisecond * time.Duration(ms+1))
			id, err := buf.getId(ctx)
			assert.Nil(t, err)
			lock.Lock()
			ids = append(ids, id)
			lock.Unlock()
		}(t)
	}

	wg.Wait()
	assert.EqualValues(t, false, misc.HasDupElems(ids))
	// t.Logf("ids = %v\n", ids)
}
