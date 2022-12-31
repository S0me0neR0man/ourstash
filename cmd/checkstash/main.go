package main

import (
	"context"
	"fmt"
	"log"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"ourstash/internal/grpcproto"
)

const (
	insertInterval = time.Millisecond * 10
)

type oneRecord struct {
	guid string
	data map[string]any

	deleted bool
}

func main() {
	conn, err := grpc.Dial(":3200", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	c := grpcproto.NewStashClient(conn)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)
	defer stop()

	var wg sync.WaitGroup
	wg.Add(1)
	go func(ctx context.Context) {
		defer wg.Done()
		for {
			select {
			case <-ctx.Done():
				return
			default:
				log.Print(".")
				time.Sleep(insertInterval)
			}
		}
	}(ctx)

	insert(c)
	wg.Wait()
}

func insert(c grpcproto.StashClient) {
	resp, err := c.Get(context.Background(), &grpcproto.GetRequest{
		Section: 0,
		Guid:    "test",
	})
	if err != nil {
		log.Fatal(err)
	}
	if resp.Error != "" {
		fmt.Println(resp.Error)
	}
	log.Println(resp.Data)
}
