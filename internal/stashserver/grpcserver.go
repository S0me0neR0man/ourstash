package stashserver

import (
	"context"
	"errors"
	"fmt"
	"net"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/anypb"

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

func (ss *StashServer) Get(ctx context.Context, in *grpcproto.GetRequest) (*grpcproto.GetResponse, error) {
	var resp grpcproto.GetResponse

	section, err := ss.getSection(in.Section)
	if err != nil {
		resp.Error = err.Error()
		return &resp, nil
	}

	var data map[string]any
	data, err = ss.stash.Get(section, stashdb.GUIDType(in.GetGuid()))
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

func (ss *StashServer) Update(ctx context.Context, in *grpcproto.UpdateRequest) (*grpcproto.UpdateResponse, error) {
	var resp grpcproto.UpdateResponse

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
	err = ss.stash.Update(section, stashdb.GUIDType(in.Guid), data)
	if err != nil {
		resp.Error = err.Error()
	}

	return &resp, nil
}

func (ss *StashServer) Remove(ctx context.Context, in *grpcproto.RemoveRequest) (*grpcproto.RemoveResponse, error) {
	var resp grpcproto.RemoveResponse

	section, err := ss.getSection(in.Section)
	if err != nil {
		resp.Error = err.Error()
		return &resp, nil
	}

	err = ss.stash.Remove(section, stashdb.GUIDType(in.GetGuid()))
	if err != nil {
		resp.Error = err.Error()
	}
	if err != nil && !errors.Is(err, stashdb.ErrRecordNotFound) {
		return &resp, err
	}

	return &resp, nil
}

func (ss *StashServer) getSection(in uint32) (stashdb.SectionIdType, error) {
	if in == 0 || in > 254 {
		return 0xff, fmt.Errorf("section must be in [1 ... 254]")
	}
	return stashdb.SectionIdType(in), nil
}

func (ss *StashServer) toStashMap(in map[string]*anypb.Any) (map[string]any, error) {
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
