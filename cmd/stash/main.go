package main

import (
	"context"
	"log"
	"os/signal"
	"syscall"

	"go.uber.org/zap"

	"ourstash/internal/config"
	"ourstash/internal/server"
	"ourstash/internal/stashdb"
)

func main() {
	logger, err := zap.NewProduction() // or NewProduction, or NewDevelopment
	if err != nil {
		log.Fatal(err)
	}
	logger.Info("stash server starting on port 3200 ...")

	var stash *stashdb.Stash
	conf := config.NewConfig()
	stash, err = stashdb.NewStash(conf, logger)
	if err != nil {
		logger.Sugar().Fatalw("Fatal error", "err", err)
	}
	server := server.NewStashServer(stash, conf, logger)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)
	defer stop()

	if err := server.Start(ctx); err != nil {
		log.Println(err)
	}

	server.Wait()
}
