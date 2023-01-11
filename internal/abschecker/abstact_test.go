package abschecker

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"reflect"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func check(before any, after any) error {
	if !reflect.DeepEqual(before, after) {
		return fmt.Errorf("check error %v not equal %v", before, after)
	}
	return nil
}

func TestStateRouter_Go(t *testing.T) {
	logger, err := zap.NewDevelopment()
	require.NoError(t, err)

	var super *StateSupervisor
	super, err = NewStateSupervisor(logger)
	require.NoError(t, err)
	require.NotNil(t, super)

	create := NewState("create", 3, logger)
	create.SetCheckFunc(check)
	split := NewState("split", 3, logger)
	split.SetCheckFunc(check)
	road1 := NewState("road1", 2, logger)
	road1.SetCheckFunc(check)
	road2 := NewState("road2", 2, logger)
	road2.SetCheckFunc(check)

	create.SetDoFunc(func(data any) (any, *State, error) {
		return data, split, nil
	})

	split.SetDoFunc(func(before any) (any, *State, error) {
		switch rand.Intn(2) {
		case 0:
			return before, road1, nil
		case 1:
			return before, road2, nil
		}
		return nil, nil, errors.New("strange")
	})

	road1.SetDoFunc(func(data any) (any, *State, error) {
		i := data.(int)
		i++
		if i == 5 {
			return nil, nil, nil
		}
		return i, split, nil
	})

	road2.SetDoFunc(func(data any) (any, *State, error) {
		i := data.(int)
		i++
		if i == 5 {
			return nil, nil, nil
		}
		return i, split, nil
	})

	super.SetGenDataFunc(func() (any, *State, error) {
		return 0, create, nil
	})

	err = super.Add(create)
	require.NoError(t, err)

	err = super.Add(split)
	require.NoError(t, err)

	err = super.Add(road1)
	require.NoError(t, err)

	err = super.Add(road2)
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*60)
	defer cancel()

	err = super.Go(ctx)
	require.NoError(t, err)
}
