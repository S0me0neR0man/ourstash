package objects

import (
	"log"
	"testing"

	"github.com/stretchr/testify/require"

	"ourstash/internal/storage"
)

func TestNewTagsUnit(t *testing.T) {
	store, err := storage.NewStorage()
	require.NoError(t, err)
	require.NotNil(t, store)

	var tags *TagsUnit
	tags, err = NewTagsUnit("#one", "#two")
	require.NoError(t, err)
	require.NotNil(t, tags)
	log.Println("test new tags unit", tags)

	chainTo, err := storage.NewPutChain()
	require.NoError(t, err)
	require.NotNil(t, chainTo)
	chainTo.Attach(tags.PutMiddleware)
}
