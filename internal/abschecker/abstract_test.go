package abschecker

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"os"
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

	create := NewState("create", 2, logger)
	create.SetCheckFunc(check)
	split := NewState("split", 3, logger)
	split.SetCheckFunc(check)
	road1 := NewState("road1", 2, logger)
	road1.SetCheckFunc(check)
	road2 := NewState("road2", 2, logger)
	road2.SetCheckFunc(check)

	create.SetDoFunc(func(data any) (any, *State, error) {
		fmt.Fprint(os.Stdout, "C")
		return data, split, nil
	})

	split.SetDoFunc(func(before any) (any, *State, error) {
		fmt.Fprint(os.Stdout, "S")
		switch rand.Intn(2) {
		case 0:
			return before, road1, nil
		case 1:
			return before, road2, nil
		}
		return nil, nil, errors.New("strange")
	})

	road1.SetDoFunc(func(data any) (any, *State, error) {
		fmt.Fprint(os.Stdout, "1")
		return nil, nil, nil
	})

	road2.SetDoFunc(func(data any) (any, *State, error) {
		fmt.Fprint(os.Stdout, "2")
		return nil, nil, nil
	})

	super.SetGenDataFunc(func() (any, *State, error) {
		fmt.Fprint(os.Stdout, ">")
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

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	err = super.Go(ctx)
	require.NoError(t, err)

	// todo: пока не работает, баг в момент shutdown блокируется  иногда
	//	super.Wait()
}
