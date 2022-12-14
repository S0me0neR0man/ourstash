package stashdb

import (
	"context"
	"strconv"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_stash_Insert(t *testing.T) {
	s := NewStash(getTestLogger())
	require.NotNil(t, s)

	to := map[string]any{
		"tag":     "#tag1",
		"text":    "sample text",
		"int_val": 100,
	}

	recGuid := s.Insert(1, to)
	require.EqualValues(t, true, recGuid != "")

	from, err := s.Get(1, recGuid)
	require.NoError(t, err)
	require.EqualValues(t, to, from)
}

func Test_stash_inGoroutines(t *testing.T) {
	s := NewStash(getTestLogger())
	require.NotNil(t, s)

	goroutinesCount := 100
	var wg sync.WaitGroup
	wg.Add(goroutinesCount * 3)

	funcInsertGet := func(i int) {
		defer wg.Done()
		to := map[string]any{
			"tag":  "#tag" + strconv.Itoa(i),
			"text": "sample text" + strconv.Itoa(i),

			"int_val" + strconv.Itoa(i): i,
		}

		recGuid := s.Insert(1, to)
		keyInsert, err := s.recordKeySFG(1, recGuid)
		require.NoError(t, err, "guid=%s err=%v", recGuid, err)
		require.EqualValues(t, true, recGuid != "")

		from, err := s.Get(1, recGuid)
		require.NoError(t, err, "guid=%s keyInsert=%s", recGuid, keyInsert)
		keyGet, _ := s.recordKeySFG(1, recGuid)
		require.EqualValues(t, to, from, "guid=%s keyInsert=%s keyGet=%s", recGuid, keyInsert, keyGet)
	}

	funcInsertGetRemove := func(i int) {
		defer wg.Done()
		to := map[string]any{
			"tag":  "#tag" + strconv.Itoa(i),
			"text": "sample text" + strconv.Itoa(i),

			"int_val" + strconv.Itoa(i): i,
		}

		recGuid := s.Insert(1, to)
		keyInsert, _ := s.recordKeySFG(1, recGuid)
		require.EqualValues(t, true, recGuid != "")

		from, err := s.Get(1, recGuid)
		require.NoError(t, err)
		keyGet, _ := s.recordKeySFG(1, recGuid)
		require.EqualValues(t, to, from, "guid=%s keyInsert=%s keyGet=%s", recGuid, keyInsert, keyGet)

		err = s.Remove(1, recGuid)
		require.NoError(t, err)

		from, err = s.Get(1, recGuid)
		require.Equal(t, ErrRecordNotFound, err, "%v", from)
	}

	funcInsertUpdateRemove := func(i int) {
		defer wg.Done()
		to := map[string]any{
			"tag":  "#tag" + strconv.Itoa(i),
			"text": "sample text" + strconv.Itoa(i),

			"int_val" + strconv.Itoa(i): i,
		}

		recGuid := s.Insert(1, to)
		keyInsert, _ := s.recordKeySFG(1, recGuid)
		require.EqualValues(t, true, recGuid != "")

		from, err := s.Get(1, recGuid)
		require.NoError(t, err)
		keyGet, _ := s.recordKeySFG(1, recGuid)
		require.EqualValues(t, to, from, "guid=%s keyInsert=%s keyGet=%s", recGuid, keyInsert, keyGet)

		to2 := map[string]any{
			"text":    "sample text" + strconv.Itoa(i),
			"int_val": i,
		}

		err = s.Update(1, recGuid, to2)
		require.NoError(t, err)

		from, err = s.Get(1, recGuid)
		require.NoError(t, err)
		require.EqualValues(t, to2, from)

		err = s.Remove(1, recGuid)
		require.NoError(t, err)

		from, err = s.Get(1, recGuid)
		require.Equal(t, ErrRecordNotFound, err, "%v", from)
	}

	for i := 0; i < goroutinesCount; i++ {
		index := i
		go funcInsertGet(index)
		go funcInsertGetRemove(index * 100)
		go funcInsertUpdateRemove(index * 10000)
	}

	wg.Wait()
}

func Test_stash_Find(t *testing.T) {
	s := NewStash(getTestLogger())
	require.NotNil(t, s)

	to := []map[string]any{
		{
			"tag":     "#tag1",
			"text":    "sample text",
			"int_val": 1,
		},
		{
			"tag":     "#tag1",
			"text":    "sample text",
			"int_val": 20,
		},
		{
			"tag":     "#tag1",
			"text":    "sample text",
			"int_val": 3,
		},
	}

	for _, m := range to {
		recGuid := s.Insert(1, m)
		require.EqualValues(t, true, recGuid != "")
	}

	ctx := context.Background()
	records, err := s.Find(ctx, 1, func(m *map[string]any) (bool, bool) {
		val, ok := (*m)["int_val"]
		if ok && val.(int) < 10 {
			return true, false
		}
		return false, false
	})

	require.NoError(t, err)
	require.EqualValues(t, 2, len(records))
}
