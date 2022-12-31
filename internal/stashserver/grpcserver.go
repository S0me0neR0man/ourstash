package stashserver

import (
	"context"
	"net"

	"go.uber.org/zap"
	"google.golang.org/grpc"

	"ourstash/internal/grpcproto"
	"ourstash/internal/stashdb"
)

type StashServer struct {
	grpcproto.UnimplementedStashServer

	stash *stashdb.Stash
	sugar *zap.SugaredLogger
	gserv *grpc.Server
}

func NewStashServer(stash *stashdb.Stash, logger *zap.Logger) *StashServer {
	return &StashServer{
		stash: stash,
		sugar: logger.Sugar(),
	}
}

func (ss *StashServer) Start() error {
	listen, err := net.Listen("tcp", "127.0.0.1:3200")
	if err != nil {
		return err
	}

	ss.gserv = grpc.NewServer()
	grpcproto.RegisterStashServer(ss.gserv, ss)
	ss.sugar.Infow("gprcserver start")

	return ss.gserv.Serve(listen)
}

func (ss *StashServer) GracefulStop() {
	ss.gserv.GracefulStop()
}

func (ss *StashServer) Insert(ctx context.Context, in *grpcproto.InsertRequest) (*grpcproto.InsertResponse, error) {
	var resp grpcproto.InsertResponse

	return &resp, nil
}

func (ss *StashServer) Get(ctx context.Context, in *grpcproto.GetRequest) (*grpcproto.GetResponse, error) {
	var resp grpcproto.GetResponse

	return &resp, nil
}

func (ss *StashServer) Update(ctx context.Context, in *grpcproto.UpdateRequest) (*grpcproto.UpdateResponse, error) {
	var resp grpcproto.UpdateResponse

	return &resp, nil
}

func (ss *StashServer) Remove(ctx context.Context, in *grpcproto.RemoveRequest) (*grpcproto.RemoveResponse, error) {
	var resp grpcproto.RemoveResponse

	return &resp, nil
}
