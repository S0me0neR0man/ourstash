package main

import (
	"context"
	"log"
	"os/signal"
	"syscall"

	"go.uber.org/zap"
	"golang.org/x/oauth2"
)

func main() {
	logger, err := zap.NewProduction()
	if err != nil {
		log.Fatal(err)
	}

	checker, err := NewChecker(logger)
	if err != nil {
		logger.Sugar().Fatalw("NewChecker", "error", err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)
	defer stop()

	checker.Go(ctx)

	err = checker.Wait()
	if err != nil {
		logger.Sugar().Errorw("checker wait", "error", err)
	}
}

// fetchToken simulates a token lookup and omits the details of proper token
// acquisition. For examples of how to acquire an OAuth2 token, see:
// https://godoc.org/golang.org/x/oauth2
func fetchToken() *oauth2.Token {
	return &oauth2.Token{
		AccessToken: "some-secret-token",
	}
}
