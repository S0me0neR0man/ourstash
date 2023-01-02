package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"syscall"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/types/known/anypb"

	"ourstash/internal/grpcproto"
	"ourstash/internal/stashdb"
)

const (
	insertInterval = time.Millisecond * 10
	displayCounter = 100
)

type oneRecord struct {
	section stashdb.SectionIdType
	guid    stashdb.GUIDType
	data    map[string]*anypb.Any
	deleted bool
	updated bool
}

func newOneRecord() oneRecord {
	i := rand.Intn(100)
	rec := oneRecord{
		section: stashdb.SectionIdType(i + 1),
		deleted: false,
		updated: false,
		data:    make(map[string]*anypb.Any),
	}

	val, err := anypb.New(&grpcproto.StringData{
		Data: "#tag" + strconv.Itoa(i),
	})
	if err != nil {
		log.Fatal(err)
	}
	rec.data["tag"] = val

	val, err = anypb.New(&grpcproto.StringData{
		Data: "sample text" + strconv.Itoa(i),
	})
	if err != nil {
		log.Fatal(err)
	}
	rec.data["text"] = val

	val, err = anypb.New(&grpcproto.IntData{
		Data: int64(i),
	})
	if err != nil {
		log.Fatal(err)
	}
	rec.data["int_value"+strconv.Itoa(i)] = val

	return rec
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

	toGet := make(chan oneRecord, 10)
	toUpdate := make(chan oneRecord, 10)
	toRemove := make(chan oneRecord, 10)

	var wg sync.WaitGroup
	wg.Add(7)

	insertFunc := func(ctx context.Context) {
		defer wg.Done()
		count := 0
		for {
			select {
			case <-ctx.Done():
				return
			default:
				rec := newOneRecord()
				rec.guid, err = insert(ctx, c, rec)
				if err != nil {
					log.Println(err)
					time.Sleep(insertInterval)
					continue
				}
				count++
				if count == displayCounter {
					count = 0
					_, _ = fmt.Fprint(os.Stdout, "I")
				}
				toGet <- rec
				time.Sleep(insertInterval)
			}
		}
	}

	getFunc := func(ctx context.Context) {
		defer wg.Done()
		count := 0
		for {
			select {
			case <-ctx.Done():
				return
			case rec := <-toGet:
				after, err := get(ctx, c, rec)
				if err != nil && !rec.deleted {
					log.Println(err)
					continue
				}
				count++
				if count == displayCounter {
					count = 0
					_, _ = fmt.Fprint(os.Stdout, "G")
				}
				if rec.deleted {
					continue
				}
				if rec.guid != after.guid {
					log.Printf("get error guid before=%s after=%s", rec.guid, after.guid)
				}
				for field, val := range rec.data {
					if val.MessageIs(after.data[field]) {
						log.Printf("get error '%s' before=%v after=%v", field, val, after.data[field])
					}
				}
				if rec.updated {
					continue
				}
				if rand.Intn(2) == 0 {
					toRemove <- rec
				} else {
					toUpdate <- rec
				}
			}
		}
	}

	updateFunc := func(ctx context.Context) {
		defer wg.Done()
		count := 0
		for {
			select {
			case <-ctx.Done():
				return
			case rec := <-toUpdate:
				err := update(ctx, c, rec)
				if err != nil && !rec.deleted {
					log.Println(err)
					continue
				}
				rec.updated = true
				toGet <- rec
				count++
				if count == displayCounter {
					count = 0
					_, _ = fmt.Fprint(os.Stdout, "U")
				}
			}
		}
	}

	removeFunc := func(ctx context.Context) {
		defer wg.Done()
		count := 0
		for {
			select {
			case <-ctx.Done():
				return
			case rec := <-toRemove:
				err := remove(ctx, c, rec)
				if err != nil && !rec.deleted {
					log.Println(err)
					continue
				}
				rec.deleted = true
				toGet <- rec
				count++
				if count == displayCounter {
					count = 0
					_, _ = fmt.Fprint(os.Stdout, "R")
				}
			}
		}
	}

	go insertFunc(ctx)
	go insertFunc(ctx)
	go getFunc(ctx)
	go getFunc(ctx)
	go getFunc(ctx)
	go updateFunc(ctx)
	go removeFunc(ctx)

	wg.Wait()
}

func get(ctx context.Context, c grpcproto.StashClient, rec oneRecord) (oneRecord, error) {
	resp, err := c.Get(ctx, &grpcproto.GetRequest{
		Section: uint32(rec.section),
		Guid:    string(rec.guid),
	})

	if err != nil {
		return oneRecord{}, fmt.Errorf("c.Get: %w", err)
	}
	if resp.Error != "" {
		return oneRecord{}, fmt.Errorf("get resp.Error: %s", resp.Error)
	}

	rec.data = resp.Data
	return rec, nil
}

func insert(ctx context.Context, c grpcproto.StashClient, rec oneRecord) (stashdb.GUIDType, error) {
	resp, err := c.Insert(ctx, &grpcproto.InsertRequest{
		Section: uint32(rec.section),
		Data:    rec.data,
	})

	if err != nil {
		return "", err
	}
	if resp.Error != "" {
		return "", fmt.Errorf("insert resp.Error: %s", resp.Error)
	}

	return stashdb.GUIDType(resp.Guid), nil
}

func update(ctx context.Context, c grpcproto.StashClient, rec oneRecord) error {
	resp, err := c.Update(ctx, &grpcproto.UpdateRequest{
		Section: uint32(rec.section),
		Guid:    string(rec.guid),
		Data:    rec.data,
	})

	if err != nil {
		return fmt.Errorf("c.Update: %w", err)
	}
	if resp.Error != "" {
		return fmt.Errorf("update resp.Error: %s", resp.Error)
	}

	return nil
}

func remove(ctx context.Context, c grpcproto.StashClient, rec oneRecord) error {
	resp, err := c.Remove(ctx, &grpcproto.RemoveRequest{
		Section: uint32(rec.section),
		Guid:    string(rec.guid),
	})

	if err != nil {
		return fmt.Errorf("c.Remove: %w", err)
	}
	if resp.Error != "" {
		return fmt.Errorf("remove resp.Error: %s", resp.Error)
	}

	return nil
}
