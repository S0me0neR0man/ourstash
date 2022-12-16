package storage

import (
	"context"
	"fmt"
	"log"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

const (
	loggerSyncInterval = time.Second
)


var (
	instance *singletonLogger
	once sync.Once
)


type singletonLogger struct {
	logger *zap.Logger
	cancel context.CancelFunc
	mu sync.RWMutex
}

func (l *singletonLogger) sync() {
	if err := l.logger.Sync(); err != nil {
		l.logger.Fatal("sync error")
	}
}

func (l *singletonLogger) StartSync() {
	if l.cancel != nil {
		l.StopSync()
	}
	ctx, cancel := context.WithCancel(context.Background())
	l.cancel = cancel

	go func(ctx context.Context) {
		ticker := time.NewTicker(loggerSyncInterval)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				l.sync()
			case <-ctx.Done():
				return
			}
		}
	}(ctx)
}

func (l *singletonLogger) StopSync() {
	if l.cancel != nil {
		l.cancel()
		l.cancel = nil
	}
}

func GetSingletonLogger() *singletonLogger {
	once.Do(func() {
		logger, err:= zap.NewDevelopment() // or NewProduction, or NewDevelopment
		if err != nil {
			log.Fatal(err)
		}
		instance = &singletonLogger{logger: logger}
	})

	return instance
}

func Test_redBlackTree_Put(t1 *testing.T) {
	slogger := GetSingletonLogger()
	defer slogger.StopSync()

	tree := NewRedBlackTree(slogger.logger)
	require.NotNil(t1, tree)
	require.EqualValues(t1, tree.Size(), 0, "not empty")

	tree.Put(NewKey(0, 1, 0, 0))
	tree.Put(NewKey(0, 2, 0, 0))
	tree.Put(NewKey(0, 1, 0, 0))
	tree.Put(NewKey(0, 3, 0, 0))
	tree.Put(NewKey(0, 4, 0, 0))
	tree.Put(NewKey(0, 5, 0, 0))
	tree.Put(NewKey(0, 6, 0, 0))

	fmt.Println(tree)

	// RedBlackTree
	// │           ┌── R 00 0000000000000006 0000 0000
	// │       ┌── B 00 0000000000000005 0000 0000
	// │   ┌── R 00 0000000000000004 0000 0000
	// │   │   └── B 00 0000000000000003 0000 0000
	// └── B 00 0000000000000002 0000 0000
	//     └── B 00 0000000000000001 0000 0000

	require.EqualValues(t1, 6, tree.Size(), "wrong size`")
	require.EqualValues(t1, 4, tree.GetNode(NewKey(0, 4, 0, 0)).Size(tree), "wrong size`")
	require.EqualValues(t1, 6, tree.GetNode(NewKey(0, 2, 0, 0)).Size(tree), "wrong size`")
	require.EqualValues(t1, 0, tree.GetNode(NewKey(0, 8, 0, 0)).Size(tree), "wrong size`")
}

func TestRedBlackTree_Remove(t1 *testing.T) {
	slogger := GetSingletonLogger()
	defer slogger.StopSync()

	tree := NewRedBlackTree(slogger.logger)
	require.NotNil(t1, tree)
	require.EqualValues(t1, tree.Size(), 0, "not empty")

	tree.Put(NewKey(0, 10, 0, 0))
	tree.Put(NewKey(0, 9, 0, 0))
	tree.Put(NewKey(0, 8, 0, 0))
	tree.Put(NewKey(0, 7, 0, 0))
	tree.Put(NewKey(0, 1, 0, 0))
	tree.Put(NewKey(0, 2, 0, 0))
	tree.Put(NewKey(0, 3, 0, 0))
	tree.Put(NewKey(0, 1, 0, 0))
	tree.Put(NewKey(0, 7, 0, 0))
	tree.Put(NewKey(0, 4, 0, 0))
	tree.Put(NewKey(0, 5, 0, 0))
	tree.Put(NewKey(0, 6, 0, 0))

	fmt.Println(tree)

	// RedBlackTree
	// │       ┌── B 00 000000000000000a 0000 0000
	// │   ┌── B 00 0000000000000009 0000 0000
	// │   │   └── B 00 0000000000000008 0000 0000
	// └── B 00 0000000000000007 0000 0000
	//     │           ┌── R 00 0000000000000006 0000 0000
	//     │       ┌── B 00 0000000000000005 0000 0000
	//     │   ┌── R 00 0000000000000004 0000 0000
	//     │   │   └── B 00 0000000000000003 0000 0000
	//     └── B 00 0000000000000002 0000 0000
	//         └── B 00 0000000000000001 0000 0000

	require.EqualValues(t1, 10, tree.Size(), "wrong size`")
	require.EqualValues(t1, 3, tree.GetNode(NewKey(0, 9, 0, 0)).Size(tree), "wrong size`")
	require.EqualValues(t1, 6, tree.GetNode(NewKey(0, 2, 0, 0)).Size(tree), "wrong size`")

	tree.Remove(NewKey(0, 10, 0, 0))
	tree.Remove(NewKey(0, 9, 0, 0))
	tree.Remove(NewKey(0, 8, 0, 0))
	tree.Remove(NewKey(0, 7, 0, 0))
	tree.Remove(NewKey(0, 9, 0, 0))
	tree.Remove(NewKey(0, 8, 0, 0))

	fmt.Println(tree)

	// RedBlackTree
	// │   ┌── B 00 0000000000000006 0000 0000
	// │   │   └── R 00 0000000000000005 0000 0000
	// └── B 00 0000000000000004 0000 0000
	//     │   ┌── B 00 0000000000000003 0000 0000
	//     └── R 00 0000000000000002 0000 0000
	//         └── B 00 0000000000000001 0000 0000

	require.EqualValues(t1, 6, tree.Size(), "wrong size`")
	require.EqualValues(t1, 3, tree.GetNode(NewKey(0, 2, 0, 0)).Size(tree), "wrong size`")
	require.EqualValues(t1, 2, tree.GetNode(NewKey(0, 6, 0, 0)).Size(tree), "wrong size`")
}

func TestRedBlackTree_inGoroutines(t1 *testing.T) {
	slogger := GetSingletonLogger()
	defer slogger.StopSync()
	sugar := slogger.logger.Sugar()

	tree := NewRedBlackTree(slogger.logger)
	tree.Debug = true
	require.NotNil(t1, tree)
	require.EqualValues(t1, tree.Size(), 0, "not empty")

	goroutinesCount := 5
	var wg sync.WaitGroup
	wg.Add(goroutinesCount *3)

	func1 := func() {
		tree.Put(NewKey(0, 10, 0, 0))
		tree.Put(NewKey(0, 9, 0, 0))
		tree.Put(NewKey(0, 8, 0, 0))
		tree.Put(NewKey(0, 7, 0, 0))
		tree.Put(NewKey(0, 1, 0, 0))
		tree.Put(NewKey(0, 2, 0, 0))
		tree.Put(NewKey(0, 3, 0, 0))
		tree.Put(NewKey(0, 1, 0, 0))
		tree.Put(NewKey(0, 7, 0, 0))
		tree.Put(NewKey(0, 4, 0, 0))
		tree.Put(NewKey(0, 5, 0, 0))
		tree.Put(NewKey(0, 6, 0, 0))
		wg.Done()
	}
	sugar.Infow("func1", "func1", fmt.Sprintf("%p", func1))

	func2 := func() {
		tree.Remove(NewKey(0, 10, 0, 0))
		tree.Remove(NewKey(0, 9, 0, 0))
		tree.Remove(NewKey(0, 8, 0, 0))
		tree.Remove(NewKey(0, 7, 0, 0))
		tree.Remove(NewKey(0, 9, 0, 0))
		tree.Remove(NewKey(0, 8, 0, 0))
		wg.Done()
	}
	sugar.Infow("func2", "func2", fmt.Sprintf("%p", func2))

	func3 := func() {
		it := tree.Iterator()
		it.Begin()
		flag := true
		for flag {
			flag = it.Next()
		}
		wg.Done()
	}
	sugar.Infow("func3", "func3", fmt.Sprintf("%p", func3))

	func4 := func() {
		it := tree.Iterator()
		it.End()
		flag := true
		for flag {
			flag = it.Prev()
		}
		wg.Done()
	}
	sugar.Infow("func4", "func4", fmt.Sprintf("%p", func4))

	for i := 0; i < goroutinesCount; i++ {
		go func1()
		go func2()
		//go func3()
		go func4()
	}

	wg.Wait()
}
