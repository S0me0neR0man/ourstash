package storage

import (
	"log"
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_stash_fields(t *testing.T) {
	s := newStash(getTestLogger())
	require.NotNil(t, s)

	f := s.fields(1)
	require.NotNil(t, f)

	to := map[string]any{
		"tag":     "#tag1",
		"text":    "sample text",
		"int_val": 100,
	}

	rec := s.Insert(1, to)

	log.Println(s)
	fields := s.fields(1)
	log.Println(fields)
	require.Less(t, RecordIdType(0), rec)

	from, err := s.Get(1, rec)
	require.NoError(t, err)
	require.EqualValues(t, to, from)
}
