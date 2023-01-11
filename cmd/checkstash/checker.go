package main

import (
	"context"
	"fmt"
	"math/rand"
	"os"
	"strconv"
	"sync"

	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/anypb"

	"github.com/S0me0neR0man/ourstash/internal/client"
	"github.com/S0me0neR0man/ourstash/internal/grpcproto"
)

const (
	displayCounter = 100
)

type simpleRecord struct {
	section  uint32
	guid     string
	data     client.GRPCData
	deleted  bool
	updated  bool
	replaced bool
}

func (s simpleRecord) String() string {
	if s.guid == "" {
		return "uninitialized"
	}
	return fmt.Sprintf("guid=%s deleted=%v updated=%v data=%v", s.guid, s.deleted, s.updated, s.data)
}

func newSimpleRecord() (simpleRecord, error) {
	i := rand.Intn(100)
	rec := simpleRecord{
		section: uint32(i + 1),
		deleted: false,
		updated: false,
		data:    make(client.GRPCData),
	}

	val, err := anypb.New(&grpcproto.StringData{
		Data: "#tag" + strconv.Itoa(i),
	})
	if err != nil {
		return simpleRecord{}, err
	}
	rec.data["tag"] = val

	val, err = anypb.New(&grpcproto.StringData{
		Data: "sample text" + strconv.Itoa(i),
	})
	if err != nil {
		return simpleRecord{}, err
	}
	rec.data["text"] = val

	val, err = anypb.New(&grpcproto.IntData{
		Data: int64(i),
	})
	if err != nil {
		return simpleRecord{}, err
	}
	rec.data["int_value"+strconv.Itoa(i)] = val

	return rec, nil
}

// TODO: разделить обхода (граф) и логику взаимодействия с объектом тестирования

type Checker struct {
	toDisplay chan string

	toGet      chan simpleRecord
	toUpdate   chan simpleRecord
	toRemove   chan simpleRecord
	toReplace  chan simpleRecord
	toGetAfter chan simpleRecord

	wg sync.WaitGroup

	client *client.GRPCClient
	sugar  *zap.SugaredLogger
}

func NewChecker(logger *zap.Logger) (*Checker, error) {
	c, err := client.NewGRPClient()
	if err != nil {
		return nil, err
	}

	return &Checker{
		client:     c,
		sugar:      logger.Sugar(),
		toDisplay:  make(chan string),
		toGet:      make(chan simpleRecord),
		toUpdate:   make(chan simpleRecord),
		toReplace:  make(chan simpleRecord),
		toRemove:   make(chan simpleRecord),
		toGetAfter: make(chan simpleRecord),
	}, nil
}

func (c *Checker) Go(ctx context.Context) {
	c.wg.Add(16)

	go c.display(ctx)
	go c.display(ctx)

	go c.insert(ctx)
	go c.insert(ctx)
	go c.get(ctx)
	go c.get(ctx)
	go c.getAfter(ctx)
	go c.getAfter(ctx)
	go c.get(ctx)
	go c.get(ctx)
	go c.update(ctx)
	go c.update(ctx)
	go c.remove(ctx)
	go c.remove(ctx)
	go c.replace(ctx)
	go c.replace(ctx)
}

func (c *Checker) Wait() error {
	c.wg.Wait()
	return c.client.Close()
}

func (c *Checker) display(ctx context.Context) {
	defer c.wg.Done()
	c.sugar.Infow("display start")

	for {
		select {
		case <-ctx.Done():
			c.sugar.Infow("display done")
			return
		case s := <-c.toDisplay:
			c.sugar.Debugw("toDisplay", "s", s)
			n, err := fmt.Fprint(os.Stdout, s)
			if err != nil {
				c.sugar.Fatalw("fprintf stdout", "err", err, "n", n)
			}
		}
	}
}

func (c *Checker) insert(ctx context.Context) {
	const msg = "insert"
	defer c.wg.Done()

	count := 0
	c.sugar.Infow(msg + " start")

	for {
		select {
		case <-ctx.Done():
			c.sugar.Infow(msg + " done")
			return
		default:
			c.sugar.Debugw(msg + " default")
			rec, err := newSimpleRecord()
			if err != nil {
				c.sugar.Fatalw(msg, "error", err)
			}
			rec.guid, err = c.client.Insert(ctx, rec.section, rec.data)
			if err != nil {
				c.sugar.Fatalw(msg, "error", err)
			}
			c.sugar.Debugw(msg+" ok", "rec", rec)

			c.toGet <- rec
			c.sugar.Debugw(msg+" toGet<-", "rec", rec)

			count++
			if count == displayCounter {
				count = 0
				c.toDisplay <- "I"
			}
		}
	}
}

func (c *Checker) get(ctx context.Context) {
	const msg = "get"
	defer c.wg.Done()

	c.sugar.Infow(msg + " start")
	count := 0

	for {
		select {
		case <-ctx.Done():
			c.sugar.Infow(msg + " done")
			return

		case rec := <-c.toGet:
			c.sugar.Debugw(msg+" <-toGet", "rec", rec)
			getData, err := c.client.Get(ctx, rec.guid)
			if err != nil {
				c.sugar.Errorw(msg, "error", err)
				continue
			}
			c.sugar.Debugw(msg+" from stash ok", "guid", rec.guid, "data", getData)

			c.compare(rec.guid, rec.data, getData)
			c.sugar.Debugw(msg+" compare done", "rec", rec)

			count++
			if count == displayCounter {
				count = 0
				c.toDisplay <- "G"
			}

			switch rand.Intn(3) {
			case 0:
				c.sugar.Debugw(msg+" before toRemote<-", "rec", rec)
				c.toRemove <- rec
				c.sugar.Debugw(msg+" toRemove<- ok", "rec", rec)
			case 1:
				c.sugar.Debugw(msg+" before toUpdate<-", "rec", rec)
				c.toUpdate <- rec
				c.sugar.Debugw(msg+" toUpdate<- ok", "rec", rec)
			case 2:
				c.sugar.Debugw(msg+" before toReplace<-", "rec", rec)
				c.toReplace <- rec
				c.sugar.Debugw(msg+" toUpdate<- ok", "rec", rec)
			}
		}
	}
}

func (c *Checker) update(ctx context.Context) {
	const msg = "update"
	defer c.wg.Done()

	c.sugar.Infow(msg + " start")
	count := 0

	for {
		select {
		case <-ctx.Done():
			c.sugar.Infow(msg + " done")
			return
		case rec := <-c.toUpdate:
			c.sugar.Debugw(msg+" <-toUpdate ok", "rec", rec)
			err := c.client.Update(ctx, rec.guid, rec.data)
			if err != nil {
				c.sugar.Errorw(msg, "error", err)
			}

			rec.updated = true
			c.sugar.Debugw(msg+" ok", "rec", rec)

			c.toGetAfter <- rec
			c.sugar.Debugw(msg+" toGetAfter<- ok", "rec", rec)

			count++
			if count == displayCounter {
				count = 0
				c.toDisplay <- "U"
			}
		}
	}
}

func (c *Checker) remove(ctx context.Context) {
	const msg = "remove"
	defer c.wg.Done()

	c.sugar.Infow(msg + " start")
	count := 0

	for {
		select {
		case <-ctx.Done():
			c.sugar.Infow(msg + " done")
			return
		case rec := <-c.toRemove:
			c.sugar.Debugw(msg+" <-toRemove ok", "rec", rec)
			err := c.client.Remove(ctx, rec.guid)
			if err != nil {
				c.sugar.Errorw(msg, "error", err)
			}

			rec.deleted = true
			c.sugar.Debugw(msg+" ok", "rec", rec)

			c.toGetAfter <- rec
			c.sugar.Debugw(msg+" toGetAfter<- ok", "rec", rec)

			count++
			if count == displayCounter {
				count = 0
				c.toDisplay <- "R"
			}
		}
	}
}

func (c *Checker) getAfter(ctx context.Context) {
	const msg = "getAfter"
	defer c.wg.Done()

	c.sugar.Infow(msg + " start")
	count := 0

	for {
		select {
		case <-ctx.Done():
			c.sugar.Infow(msg + " done")
			return

		case rec := <-c.toGetAfter:
			c.sugar.Debugw(msg+" <-toGetAfter", "rec", rec)
			getData, err := c.client.Get(ctx, rec.guid)
			if err != nil && !rec.deleted {
				c.sugar.Errorw(msg, "error", err)
				continue
			}
			c.sugar.Debugw(msg+" from stash ok", "guid", rec.guid, "data", getData)

			if !rec.deleted {
				c.compare(rec.guid, rec.data, getData)
			}
			c.sugar.Debugw(msg+" compare done", "rec", rec)

			count++
			if count == displayCounter {
				count = 0
				c.toDisplay <- "A"
			}
		}
	}
}

func (c *Checker) compare(guid string, before client.GRPCData, after client.GRPCData) {
	for field, val := range before {
		if val.MessageIs(after[field]) {
			c.sugar.Errorw("not equal", "guid", guid, "field", field, "before", val, "after", after[field])
		}
	}
	if len(before) != len(after) {
		c.sugar.Errorw("wrong length", "guid", guid, "before", len(before), "after", len(after))
	}
}

func (c *Checker) replace(ctx context.Context) {
	const msg = "replace"
	defer c.wg.Done()

	c.sugar.Infow(msg + " start")
	count := 0

	for {
		select {
		case <-ctx.Done():
			c.sugar.Infow(msg + " done")
			return
		case rec := <-c.toReplace:
			c.sugar.Debugw(msg+" <-toReplace ok", "rec", rec)
			err := c.client.Replace(ctx, rec.guid, rec.data)
			if err != nil {
				c.sugar.Errorw(msg, "error", err)
			}

			rec.replaced = true
			c.sugar.Debugw(msg+" ok", "rec", rec)

			c.toGetAfter <- rec
			c.sugar.Debugw(msg+" toGetAfter<- ok", "rec", rec)

			count++
			if count == displayCounter {
				count = 0
				c.toDisplay <- "E"
			}
		}
	}
}
