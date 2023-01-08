package stashdb

import (
	"context"
	"math/rand"
	"os"
	"strconv"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"

	"ourstash/internal/config"
)

const (
	countRecords        = 100
	countFieldsInRecord = 10
)

var (
	onceConf sync.Once
	conf     *config.Config
)

func getConfig() *config.Config {
	onceConf.Do(func() {
		_ = os.Setenv("STORE_FILE", "db/test_stash.data")
		conf = config.NewConfig()
	})
	return conf
}

func Test_stash_Insert(t *testing.T) {
	conf := getConfig()
	conf.Restore = false
	s, err := NewStash(conf, getTestLogger())
	require.NoError(t, err)
	require.NotNil(t, s)

	to := map[string]any{
		"tag":     "#tag1",
		"text":    "sample text",
		"int_val": 100,
	}

	recGuid := s.Insert(1, to)
	require.EqualValues(t, true, recGuid != "")

	from, err := s.Get(recGuid)
	require.NoError(t, err)
	require.EqualValues(t, to, from)
}

func Test_stash_inGoroutines(t *testing.T) {
	conf := getConfig()
	conf.Restore = false
	s, err := NewStash(conf, getTestLogger())
	require.NoError(t, err)
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
		keyInsert, err := s.recordKeySFG(recGuid)
		require.NoError(t, err, "Guid=%s err=%v", recGuid, err)
		require.EqualValues(t, true, recGuid != "")

		from, err := s.Get(recGuid)
		require.NoError(t, err, "Guid=%s keyInsert=%s", recGuid, keyInsert)
		keyGet, _ := s.recordKeySFG(recGuid)
		require.EqualValues(t, to, from, "Guid=%s keyInsert=%s keyGet=%s", recGuid, keyInsert, keyGet)
	}

	funcInsertGetRemove := func(i int) {
		defer wg.Done()
		to := map[string]any{
			"tag":  "#tag" + strconv.Itoa(i),
			"text": "sample text" + strconv.Itoa(i),

			"int_val" + strconv.Itoa(i): i,
		}

		recGuid := s.Insert(1, to)
		keyInsert, _ := s.recordKeySFG(recGuid)
		require.EqualValues(t, true, recGuid != "")

		from, err := s.Get(recGuid)
		require.NoError(t, err)
		keyGet, _ := s.recordKeySFG(recGuid)
		require.EqualValues(t, to, from, "Guid=%s keyInsert=%s keyGet=%s", recGuid, keyInsert, keyGet)

		err = s.Remove(recGuid)
		require.NoError(t, err)

		from, err = s.Get(recGuid)
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
		keyInsert, _ := s.recordKeySFG(recGuid)
		require.EqualValues(t, true, recGuid != "")

		from, err := s.Get(recGuid)
		require.NoError(t, err)
		keyGet, _ := s.recordKeySFG(recGuid)
		require.EqualValues(t, to, from, "Guid=%s keyInsert=%s keyGet=%s", recGuid, keyInsert, keyGet)

		to2 := map[string]any{
			"text":    "sample text" + strconv.Itoa(i),
			"int_val": i,
		}

		err = s.Update(recGuid, to2)
		require.NoError(t, err)

		from, err = s.Get(recGuid)
		require.NoError(t, err)
		require.EqualValues(t, to2, from)

		err = s.Remove(recGuid)
		require.NoError(t, err)

		from, err = s.Get(recGuid)
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
	s, err := NewStash(getConfig(), getTestLogger())
	require.NoError(t, err)
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

func generateTestData() []map[string]any {
	ret := make([]map[string]any, 0)
	for i := 0; i < countRecords; i++ {
		ret = append(ret, make(map[string]any))
		for k := 0; k < countFieldsInRecord; k++ {
			switch rand.Intn(2) {
			case 0:
				ret[i]["intVal_"+strconv.Itoa(k)] = k
			case 1:
				ret[i]["stringVal_"+strconv.Itoa(k)] = "test string " + strconv.Itoa(k)
			}
		}
	}
	return ret
}

func TestStash_SaveToDisk_LoadFromDisk(t *testing.T) {
	conf := getConfig()
	conf.Restore = false
	conf.StoreInterval = 0
	sTo, err := NewStash(conf, getTestLogger())
	require.NotNil(t, sTo)

	to := generateTestData()

	for _, m := range to {
		recGuid := sTo.Insert(1, m)
		require.EqualValues(t, true, recGuid != "")
	}

	ctx := context.Background()
	err = sTo.SaveToDisk(ctx)
	require.NoError(t, err)

	conf.Restore = true
	var sFrom *Stash
	sFrom, err = NewStash(conf, getTestLogger())
	require.NotNil(t, sFrom)
	require.EqualValues(t, sTo.size, sFrom.size)

	controlTo := sTo.copyData(ctx)
	controlFrom := sFrom.copyData(ctx)
	require.Equal(t, controlTo, controlFrom)
}
