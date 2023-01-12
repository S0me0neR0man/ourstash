package abschecker

import (
	"context"
	"fmt"
	"math/rand"
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

const (
	createState = "create"
	splitState  = "split"
	road1State  = "road1"
	road2State  = "road2"
)

func check(before DataToBeVerified, after DataToBeVerified) error {
	if !reflect.DeepEqual(before.Data, after.Data) {
		return fmt.Errorf("check error: %v not equal %v", before, after)
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

	create := NewState(createState, 1, logger)
	create.SetCheckFunc(check)
	split := NewState(splitState, 1, logger)
	split.SetCheckFunc(check)
	road1 := NewState(road1State, 1, logger)
	road1.SetCheckFunc(check)
	road2 := NewState(road2State, 1, logger)
	road2.SetCheckFunc(check)

	create.SetDoFunc(func(data DataToBeVerified) (DataToBeVerified, error) {
		fmt.Fprint(os.Stdout, "C")
		data.NextState = splitState
		return data, nil
	})

	split.SetDoFunc(func(data DataToBeVerified) (DataToBeVerified, error) {
		fmt.Fprint(os.Stdout, "S")

		switch rand.Intn(2) {
		case 0:
			data.NextState = road1State
		case 1:
			data.NextState = road2State
		}
		return data, nil
	})

	road1.SetDoFunc(func(data DataToBeVerified) (DataToBeVerified, error) {
		fmt.Fprint(os.Stdout, "1")
		return data, nil
	})

	road2.SetDoFunc(func(data DataToBeVerified) (DataToBeVerified, error) {
		fmt.Fprint(os.Stdout, "2")
		return data, nil
	})

	super.SetGenDataFunc(func() (DataToBeVerified, error) {
		fmt.Fprint(os.Stdout, ">")
		return DataToBeVerified{
			CurrentState: createState,
			Data:         101,
		}, nil
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

	super.Wait()
}
