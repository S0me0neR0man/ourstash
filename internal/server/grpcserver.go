package server

import (
	"context"
	"errors"
	"fmt"
	"net"
	"sync"
	"time"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/anypb"

	"github.com/S0me0neR0man/ourstash/internal/config"
	"github.com/S0me0neR0man/ourstash/internal/grpcproto"
	"github.com/S0me0neR0man/ourstash/internal/stashdb"
)

var (
	errMissingMetadata = status.Errorf(codes.InvalidArgument, "missing metadata")
	//errInvalidToken    = status.Errorf(codes.Unauthenticated, "invalid token")
)

const ()

type GRPCServer struct {
	grpcproto.UnimplementedStashServer

	stash *stashdb.Stash
	sugar *zap.SugaredLogger
	gserv *grpc.Server
	conf  *config.Config

	wg sync.WaitGroup
}

func NewStashServer(stash *stashdb.Stash, conf *config.Config, logger *zap.Logger) *GRPCServer {
	return &GRPCServer{
		stash: stash,
		conf:  conf,
		sugar: logger.Sugar(),
	}
}

func (ss *GRPCServer) Start(ctx context.Context) error {
	opts := []grpc.ServerOption{
		grpc.UnaryInterceptor(ss.ensureValidToken),
	}

	ss.gserv = grpc.NewServer(opts...)
	grpcproto.RegisterStashServer(ss.gserv, ss)
	ss.sugar.Infow("gprcserver start")

	lis, err := net.Listen("tcp", "127.0.0.1:3200")
	if err != nil {
		return err
	}

	go ss.saveToDisk(ctx)
	go ss.gracefulStop(ctx)

	return ss.gserv.Serve(lis)
}

func (ss *GRPCServer) saveToDisk(ctx context.Context) {
	if ss.conf.StoreInterval == 0 {
		return
	}

	ss.wg.Add(1)
	defer ss.wg.Done()

	ticker := time.NewTicker(ss.conf.StoreInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			err := ss.stash.SaveToDisk(ctx)
			if err != nil {
				ss.sugar.Errorw("stash.SaveToDisk", "error", err)
			}
		case <-ctx.Done():
			return
		}
	}
}

func (ss *GRPCServer) gracefulStop(ctx context.Context) {
	ss.wg.Add(1)
	defer ss.wg.Done()

	<-ctx.Done()
	ss.gserv.GracefulStop()
}

func (ss *GRPCServer) Wait() {
	ss.wg.Wait()
}

func (ss *GRPCServer) ensureValidToken(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, errMissingMetadata
	}
	ss.sugar.Debugw("ensureValidToken", "token", md["authorization"])

	// The keys within metadata.MD are normalized to lowercase.
	// See: https://godoc.org/google.golang.org/grpc/metadata#New
	//if !valid(md["authorization"]) {
	//	return nil, errInvalidToken
	//}
	// Continue execution of handler after ensuring a valid token.
	return handler(ctx, req)
}

func (ss *GRPCServer) Insert(ctx context.Context, in *grpcproto.InsertRequest) (*grpcproto.InsertResponse, error) {
	var resp grpcproto.InsertResponse

	section, err := ss.getSection(in.Section)
	if err != nil {
		resp.Error = err.Error()
		return &resp, nil
	}

	var data map[string]any
	data, err = ss.toStashMap(in.Data)
	if err != nil {
		resp.Error = err.Error()
		return nil, err
	}
	resp.Guid = string(ss.stash.Insert(section, data))

	return &resp, nil
}

func (ss *GRPCServer) Get(ctx context.Context, in *grpcproto.GetRequest) (*grpcproto.GetResponse, error) {
	var resp grpcproto.GetResponse

	data, err := ss.stash.Get(stashdb.GUIDType(in.GetGuid()))
	if err != nil {
		resp.Error = err.Error()
		return &resp, nil
	}

	resp.Data = make(map[string]*anypb.Any)
	for key, val := range data {
		switch i := val.(type) {
		case int64:
			a, err := anypb.New(&grpcproto.IntData{
				Data: val.(int64),
			})
			if err != nil {
				resp.Error = err.Error()
				return &resp, err
			}
			resp.Data[key] = a
		case string:
			a, err := anypb.New(&grpcproto.StringData{
				Data: val.(string),
			})
			if err != nil {
				resp.Error = err.Error()
				return &resp, err
			}
			resp.Data[key] = a
		default:
			ss.sugar.Warnw("get unsupported type", "value", i)
		}
	}
	return &resp, nil
}

func (ss *GRPCServer) Update(ctx context.Context, in *grpcproto.UpdateRequest) (*grpcproto.UpdateResponse, error) {
	var resp grpcproto.UpdateResponse

	data, err := ss.toStashMap(in.Data)
	if err != nil {
		resp.Error = err.Error()
		return nil, err
	}
	err = ss.stash.Update(stashdb.GUIDType(in.Guid), data)
	if err != nil {
		resp.Error = err.Error()
	}

	return &resp, nil
}

func (ss *GRPCServer) Remove(ctx context.Context, in *grpcproto.RemoveRequest) (*grpcproto.RemoveResponse, error) {
	var resp grpcproto.RemoveResponse

	err := ss.stash.Remove(stashdb.GUIDType(in.GetGuid()))
	if err != nil {
		resp.Error = err.Error()
	}
	if err != nil && !errors.Is(err, stashdb.ErrRecordNotFound) {
		return &resp, err
	}

	return &resp, nil
}

func (ss *GRPCServer) getSection(in uint32) (stashdb.SectionIdType, error) {
	if in == 0 || in > 254 {
		return 0xff, fmt.Errorf("section must be in [1 ... 254]")
	}
	return stashdb.SectionIdType(in), nil
}

func (ss *GRPCServer) toStashMap(in map[string]*anypb.Any) (map[string]any, error) {
	out := make(map[string]any)
	for field, val := range in {
		switch val.GetTypeUrl() {
		case "type.googleapis.com/grpcs.IntData":
			intData := &grpcproto.IntData{}
			if err := val.UnmarshalTo(intData); err != nil {
				return nil, err
			}
			out[field] = intData.GetData()
		case "type.googleapis.com/grpcs.StringData":
			strData := &grpcproto.StringData{}
			if err := val.UnmarshalTo(strData); err != nil {
				return nil, err
			}
			out[field] = strData.GetData()
		default:
			ss.sugar.Errorw("unknown", "TypeUrl", val.GetTypeUrl())
		}
	}

	return out, nil
}

func (ss *GRPCServer) Replace(ctx context.Context, in *grpcproto.UpdateRequest) (*grpcproto.UpdateResponse, error) {
	var resp grpcproto.UpdateResponse

	data, err := ss.toStashMap(in.Data)
	if err != nil {
		resp.Error = err.Error()
		return nil, err
	}
	err = ss.stash.Replace(stashdb.GUIDType(in.Guid), data)
	if err != nil {
		resp.Error = err.Error()
	}

	return &resp, nil
}
