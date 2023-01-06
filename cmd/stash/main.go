package main

import (
	"context"
	"log"
	"os/signal"
	"sync"
	"syscall"

	"go.uber.org/zap"

	"ourstash/internal/stashdb"
	"ourstash/internal/stashserver"
)

func main() {
	logger, err := zap.NewDevelopment() // or NewProduction, or NewDevelopment
	if err != nil {
		log.Fatal(err)
	}
	stash := stashdb.NewStash(logger)
	s := stashserver.NewStashServer(stash, logger)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)
	defer stop()

	var wg sync.WaitGroup
	wg.Add(1)
	go func(ctx context.Context) {
		defer wg.Done()
		<-ctx.Done()
		s.GracefulStop()
	}(ctx)

	if err := s.Start(); err != nil {
		log.Println(err)
	}

	wg.Wait()
}
