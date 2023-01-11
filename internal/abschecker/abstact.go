package abschecker

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"go.uber.org/zap"
)

var (
	ErrNotInitialized = errors.New("not initialized")
	ErrStateNotFound  = errors.New("state not found")
)

type DoFunc func(any) (any, *State, error)
type CheckFunc func(any, any) error
type GenDataFunc func() (any, *State, error)

type State struct {
	Id      string
	GoCount uint

	doFunc    DoFunc
	checkFunc CheckFunc
	inChan    chan any

	sugar *zap.SugaredLogger
}

func NewState(id string, goCount uint, logger *zap.Logger) *State {
	return &State{
		Id:      id,
		GoCount: goCount,
		sugar:   logger.Sugar(),
		inChan:  make(chan any),
	}
}

func (s *State) String() string {
	return fmt.Sprintf("Id=%s gouroutines=%d", s.Id, s.GoCount)
}

func (s *State) SetDoFunc(f DoFunc) {
	s.doFunc = f
}

func (s *State) SetCheckFunc(f CheckFunc) {
	s.checkFunc = f
}

func (s *State) goWork(ctx context.Context, done chan int) {
	defer func() {
		done <- 0
	}()

	for {
		select {
		case <-ctx.Done():
			s.sugar.Infow(s.Id + " done")
			return
		case before := <-s.inChan:
			s.sugar.Errorw("goWork", "Id", s.Id, "before", before)
			after, next, err := s.doFunc(before)
			if err != nil {
				s.sugar.Errorw("doFunc", "Id", s.Id, "error", err)
				return
			}
			s.sugar.Errorw("goWork", "Id", s.Id, "after", after)
			err = nil
			if s.checkFunc == nil {
				err = s.checkFunc(before, after)
			}
			if err != nil {
				s.sugar.Errorw("checkFunc", "Id", s.Id, "error", err)
				continue
			}
			if next != nil {
				next.Push(after)
			}
		}
	}
}

func (s *State) Push(data any) {
	s.sugar.Errorw("Push", "Id", s.Id, "data", data)
	s.inChan <- data
}

type StateSupervisor struct {
	mu sync.RWMutex

	states      map[string]*State
	genDataFunc GenDataFunc

	done chan int
	wg   sync.WaitGroup

	superDone chan int
	superWG   sync.WaitGroup

	sugar *zap.SugaredLogger
}

func NewStateSupervisor(logger *zap.Logger) (*StateSupervisor, error) {
	return &StateSupervisor{
		sugar:     logger.Sugar(),
		done:      make(chan int),
		superDone: make(chan int),
		states:    make(map[string]*State),
	}, nil
}

func (sv *StateSupervisor) Add(s *State) error {
	sv.mu.Lock()
	defer sv.mu.Unlock()

	sv.states[s.Id] = s
	return nil
}

func (sv *StateSupervisor) Go(ctx context.Context) error {
	if sv.genDataFunc == nil || len(sv.states) == 0 {
		return ErrNotInitialized
	}
	sv.mu.RLock()
	go sv.startGoroutines(ctx)
	return nil
}

func (sv *StateSupervisor) SetGenDataFunc(f GenDataFunc) {
	sv.genDataFunc = f
}

func (sv *StateSupervisor) startGoroutines(ctx context.Context) {
	sv.superWG.Add(1)
	go sv.supervisor(ctx)

	sv.wg.Add(1)
	go sv.generateData(ctx)

	for id, state := range sv.states {
		if state.doFunc == nil {
			sv.sugar.Errorw("doFunc not set", "Id", id)
			continue
		}
		sv.wg.Add(1)
		go state.goWork(ctx, sv.done)
	}
}

func (sv *StateSupervisor) supervisor(ctx context.Context) {
	defer func() {
		sv.superWG.Done()
	}()

	const msg = "supervisor"
	for {
		select {
		case <-ctx.Done():
			sv.sugar.Infow(msg + " <-ctx.Done received")
		case <-sv.done:
			sv.wg.Done()
		case <-sv.superDone:
			sv.sugar.Infow(msg + " done")
			return
		}
	}
}

func (sv *StateSupervisor) Wait(ctx context.Context) {
	sv.wg.Wait()
	sv.superDone <- 0
	sv.superWG.Wait()
	sv.mu.RUnlock()
}

func (sv *StateSupervisor) generateData(ctx context.Context) {
	defer func() {
		sv.done <- 0
	}()

	msg := "generateData"
	for {
		select {
		case <-ctx.Done():
			sv.sugar.Infow(msg + " done")
			return
		default:
			data, state, err := sv.genDataFunc()
			if err != nil {
				sv.sugar.Fatalw(msg, "error", err)
				return
			}
			if state == nil {
				sv.sugar.Fatalw(msg, "error", "state is nil")
			}
			state.Push(data)
		}
	}

}
