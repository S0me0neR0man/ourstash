package storage

import (
	"log"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewStorage(t *testing.T) {
	store, err := NewStorage()
	require.NoError(t, err)
	require.NotNil(t, store)

	var tags *TagsUnit
	tags, err = NewTagsUnit("#one", "#two")
	require.NoError(t, err)
	require.NotNil(t, tags)
	log.Println(tags)

	rec := NewRecord()
	rec.UseInPut(tags.PutHandler)
	store.New(rec)

	//t.Run("new storage", func(t *testing.T) {
	//	}
	//})
}
