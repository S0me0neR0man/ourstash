package abschecker

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"go.uber.org/zap"
)

var (
	ErrNotInitialized     = errors.New("not initialized")
	ErrShutdownInProgress = errors.New("not initialized")
)

type DataToBeVerified struct {
	CurrentState string
	NextState    string
	Data         any
}

func (d DataToBeVerified) String() string {
	return fmt.Sprintf("(%s) -> (%s) %v", d.CurrentState, d.NextState, d.Data)
}

type DoFunc func(DataToBeVerified) (DataToBeVerified, error)
type CheckFunc func(DataToBeVerified, DataToBeVerified) error
type DataSourceFunc func() (DataToBeVerified, error)

type State struct {
	Id      string
	GoCount uint

	doFunc    DoFunc
	checkFunc CheckFunc
	inChan    chan DataToBeVerified

	sugar *zap.SugaredLogger
}

func NewState(id string, goCount uint, logger *zap.Logger) *State {
	return &State{
		Id:      id,
		GoCount: goCount,
		sugar:   logger.Sugar(),
		inChan:  make(chan DataToBeVerified),
	}
}

func (s *State) String() string {
	return fmt.Sprintf("%s gouroutines=%d", s.Id, s.GoCount)
}

func (s *State) SetDoFunc(f DoFunc) {
	s.doFunc = f
}

func (s *State) SetCheckFunc(f CheckFunc) {
	s.checkFunc = f
}

func (s *State) job(ctx context.Context, routerChan chan DataToBeVerified, doneChan chan any) {
	defer func() {
		doneChan <- 0
	}()

	const msg = "job"
	s.sugar.Infoln(msg, s.Id, "started")
	for {
		select {
		case <-ctx.Done():
			s.sugar.Infoln(msg, s.Id, " ctx.Done() received")
			return
		case before := <-s.inChan:
			s.sugar.Debugf("%s %s <- %v", msg, s.Id, before)
			after, err := s.doFunc(before)
			if err != nil {
				s.sugar.Debugf("%s %s doFunc: %s", msg, s.Id, err.Error())
				return
			}
			s.sugar.Debugf("%s %s after %v", msg, s.Id, after)
			err = nil
			if s.checkFunc != nil {
				err = s.checkFunc(before, after)
			}
			if err != nil {
				s.sugar.Debugf("%s %s checkFunc: %s", msg, s.Id, err.Error())
				continue
			}
			if after.NextState != "" {
				s.sugar.Debugf("%s %s -> %s", msg, s.Id, after.NextState)
				routerChan <- after
			}
		}
	}
}

func (s *State) Push(data DataToBeVerified) error {
	s.inChan <- data
	s.sugar.Debugf("Pushed %s <- %v", s.Id, data)
	return nil
}

type StateSupervisor struct {
	mu sync.RWMutex

	states map[string]*State

	sourceFunc DataSourceFunc

	done chan any
	wg   sync.WaitGroup

	routerChan chan DataToBeVerified
	routerDone chan any
	routerWG   sync.WaitGroup

	superDone chan any
	superWG   sync.WaitGroup

	sugar *zap.SugaredLogger
}

func NewStateSupervisor(logger *zap.Logger) (*StateSupervisor, error) {
	return &StateSupervisor{
		sugar:     logger.Sugar(),
		done:      make(chan any),
		superDone: make(chan any),
		states:    make(map[string]*State),
	}, nil
}

func (sv *StateSupervisor) Add(s *State) error {
	sv.mu.Lock()
	defer sv.mu.Unlock()

	if s == nil {
		return errors.New("input param is nil")
	}

	sv.states[s.Id] = s
	sv.sugar.Infof("added %v", s)
	return nil
}

func (sv *StateSupervisor) SetGenDataFunc(f DataSourceFunc) {
	sv.sourceFunc = f
}

func (sv *StateSupervisor) Go(ctx context.Context) error {
	if sv.sourceFunc == nil || len(sv.states) == 0 {
		return ErrNotInitialized
	}
	sv.sugar.Infoln("starting ...")

	sv.mu.RLock()
	sv.startGoroutines(ctx)

	return nil
}

func (sv *StateSupervisor) startGoroutines(ctx context.Context) {
	sv.superWG.Add(1)
	go sv.supervisorJob()

	sv.routerWG.Add(1)
	go sv.routeJob()

	sv.wg.Add(1)
	go sv.sourceJob(ctx)

	for id, state := range sv.states {
		if state.doFunc == nil {
			sv.sugar.Errorw("doFunc not set", "Id", id)
			continue
		}
		for i := uint(0); i < state.GoCount; i++ {
			sv.wg.Add(1)
			go state.job(ctx, sv.routerChan, sv.done)
		}
		sv.sugar.Infof("%v started", state)
	}
}

func (sv *StateSupervisor) supervisorJob() {
	defer sv.superWG.Done()

	const msg = "supervisorJob"
	sv.sugar.Infoln(msg, "started")
	for {
		select {
		case <-sv.done:
			sv.wg.Done()
		case <-sv.superDone:
			sv.sugar.Infoln(msg + " done")
			return
		}
	}
}

func (sv *StateSupervisor) routeJob() {
	defer sv.routerWG.Done()

	const msg = "routerJob"
	sv.sugar.Infoln(msg, "started")
	for {
		select {
		case <-sv.routerDone:
			sv.sugar.Infoln(msg + " done")
			return
		case data := <-sv.routerChan:
			if data.NextState == "" {
				continue
			}
			state, ok := sv.states[data.NextState]
			if !ok {
				sv.sugar.Errorf("%s error: %s not found", msg, data.NextState)
				continue
			}
			err := state.Push(data)
			if err != nil {
				sv.sugar.Errorf("%s Push error: %s", msg, err.Error())
			}
		}
	}
}

func (sv *StateSupervisor) sourceJob(ctx context.Context) {
	defer func() {
		sv.done <- 0
	}()

	const msg = "sourceJob"
	sv.sugar.Infoln(msg, "started")
	for {
		select {
		case <-ctx.Done():
			sv.sugar.Infoln(msg + " done")
			return
		default:
			data, err := sv.sourceFunc()
			if err != nil {
				sv.sugar.Fatalln(msg, err)
			}
			if data.NextState == "" {
				continue
			}
			state, ok := sv.states[data.NextState]
			if !ok {
				sv.sugar.Errorf("%s error: %s not found", msg, data.NextState)
				continue
			}
			err = state.Push(data)
			if err != nil {
				sv.sugar.Errorf("%s Push error: %s", msg, err.Error())
			}
		}
	}

}

func (sv *StateSupervisor) Wait() {
	sv.wg.Wait()

	sv.superDone <- 0
	sv.superWG.Wait()

	sv.routerDone <- 0
	sv.routerWG.Wait()

	sv.mu.RUnlock()
	sv.sugar.Infoln("ALL graceful shutdown")
}
