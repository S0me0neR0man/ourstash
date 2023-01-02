// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.2.0
// - protoc             v3.12.4
// source: internal/grpcproto/stash.proto

package grpcproto

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.32.0 or later.
const _ = grpc.SupportPackageIsVersion7

// StashClient is the client API for Stash service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type StashClient interface {
	Insert(ctx context.Context, in *InsertRequest, opts ...grpc.CallOption) (*InsertResponse, error)
	Get(ctx context.Context, in *GetRequest, opts ...grpc.CallOption) (*GetResponse, error)
	Remove(ctx context.Context, in *RemoveRequest, opts ...grpc.CallOption) (*RemoveResponse, error)
	Update(ctx context.Context, in *UpdateRequest, opts ...grpc.CallOption) (*UpdateResponse, error)
}

type stashClient struct {
	cc grpc.ClientConnInterface
}

func NewStashClient(cc grpc.ClientConnInterface) StashClient {
	return &stashClient{cc}
}

func (c *stashClient) Insert(ctx context.Context, in *InsertRequest, opts ...grpc.CallOption) (*InsertResponse, error) {
	out := new(InsertResponse)
	err := c.cc.Invoke(ctx, "/grpcs.Stash/Insert", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *stashClient) Get(ctx context.Context, in *GetRequest, opts ...grpc.CallOption) (*GetResponse, error) {
	out := new(GetResponse)
	err := c.cc.Invoke(ctx, "/grpcs.Stash/Get", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *stashClient) Remove(ctx context.Context, in *RemoveRequest, opts ...grpc.CallOption) (*RemoveResponse, error) {
	out := new(RemoveResponse)
	err := c.cc.Invoke(ctx, "/grpcs.Stash/Remove", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *stashClient) Update(ctx context.Context, in *UpdateRequest, opts ...grpc.CallOption) (*UpdateResponse, error) {
	out := new(UpdateResponse)
	err := c.cc.Invoke(ctx, "/grpcs.Stash/Update", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// StashServer is the server API for Stash service.
// All implementations must embed UnimplementedStashServer
// for forward compatibility
type StashServer interface {
	Insert(context.Context, *InsertRequest) (*InsertResponse, error)
	Get(context.Context, *GetRequest) (*GetResponse, error)
	Remove(context.Context, *RemoveRequest) (*RemoveResponse, error)
	Update(context.Context, *UpdateRequest) (*UpdateResponse, error)
	mustEmbedUnimplementedStashServer()
}

// UnimplementedStashServer must be embedded to have forward compatible implementations.
type UnimplementedStashServer struct {
}

func (UnimplementedStashServer) Insert(context.Context, *InsertRequest) (*InsertResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Insert not implemented")
}
func (UnimplementedStashServer) Get(context.Context, *GetRequest) (*GetResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Get not implemented")
}
func (UnimplementedStashServer) Remove(context.Context, *RemoveRequest) (*RemoveResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Remove not implemented")
}
func (UnimplementedStashServer) Update(context.Context, *UpdateRequest) (*UpdateResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Update not implemented")
}
func (UnimplementedStashServer) mustEmbedUnimplementedStashServer() {}

// UnsafeStashServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to StashServer will
// result in compilation errors.
type UnsafeStashServer interface {
	mustEmbedUnimplementedStashServer()
}

func RegisterStashServer(s grpc.ServiceRegistrar, srv StashServer) {
	s.RegisterService(&Stash_ServiceDesc, srv)
}

func _Stash_Insert_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(InsertRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(StashServer).Insert(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/grpcs.Stash/Insert",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(StashServer).Insert(ctx, req.(*InsertRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Stash_Get_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(StashServer).Get(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/grpcs.Stash/Get",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(StashServer).Get(ctx, req.(*GetRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Stash_Remove_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(RemoveRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(StashServer).Remove(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/grpcs.Stash/Remove",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(StashServer).Remove(ctx, req.(*RemoveRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Stash_Update_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(UpdateRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(StashServer).Update(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/grpcs.Stash/Update",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(StashServer).Update(ctx, req.(*UpdateRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// Stash_ServiceDesc is the grpc.ServiceDesc for Stash service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var Stash_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "grpcs.Stash",
	HandlerType: (*StashServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Insert",
			Handler:    _Stash_Insert_Handler,
		},
		{
			MethodName: "Get",
			Handler:    _Stash_Get_Handler,
		},
		{
			MethodName: "Remove",
			Handler:    _Stash_Remove_Handler,
		},
		{
			MethodName: "Update",
			Handler:    _Stash_Update_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "internal/grpcproto/stash.proto",
}
