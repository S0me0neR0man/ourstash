package storage

import (
	"strconv"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_stash_Insert(t *testing.T) {
	s := newStash(getTestLogger())
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
	s := newStash(getTestLogger())
	require.NotNil(t, s)

	goroutinesCount := 50
	var wg sync.WaitGroup
	wg.Add(goroutinesCount * 2)

	func1 := func(i int) {
		defer wg.Done()
		to := map[string]any{
			"tag":     "#tag" + strconv.Itoa(i),
			"text":    "sample text" + strconv.Itoa(i),
			"int_val": i,
		}

		recGuid := s.Insert(1, to)
		require.EqualValues(t, true, recGuid != "")

		from, err := s.Get(1, recGuid)
		require.NoError(t, err)
		require.EqualValues(t, to, from)
	}

	func2 := func(i int) {
		defer wg.Done()
		to := map[string]any{
			"tag":     "#tag" + strconv.Itoa(i),
			"text":    "sample text" + strconv.Itoa(i),
			"int_val": i,
		}

		recGuid := s.Insert(1, to)
		require.EqualValues(t, true, recGuid != "")

		from, err := s.Get(1, recGuid)
		require.NoError(t, err)
		require.EqualValues(t, to, from)

		err = s.Remove(1, recGuid)
		require.NoError(t, err)

		from, err = s.Get(1, recGuid)
		require.Equal(t, ErrRecordNotFound, err)
	}

	for i := 0; i < goroutinesCount; i++ {
		index := i
		go func1(index)
		go func2(index)
	}

	wg.Wait()
}
