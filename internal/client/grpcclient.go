package client

import (
	"context"
	"errors"
	"fmt"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/types/known/anypb"

	"ourstash/internal/grpcproto"
	"ourstash/internal/token"
)

type GRPCData map[string]*anypb.Any

type GRPCClient struct {
	conn   *grpc.ClientConn
	client grpcproto.StashClient
}

func NewGRPClient() (*GRPCClient, error) {
	c := GRPCClient{}

	opts := []grpc.DialOption{
		grpc.WithPerRPCCredentials(&token.Tokens{}),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	}

	var err error
	c.conn, err = grpc.Dial(":3200", opts...)
	if err != nil {
		return nil, err
	}
	c.client = grpcproto.NewStashClient(c.conn)

	return &c, nil
}

func (c *GRPCClient) Close() error {
	return c.conn.Close()
}

func (c *GRPCClient) Get(ctx context.Context, guid string) (GRPCData, error) {
	resp, err := c.client.Get(ctx, &grpcproto.GetRequest{
		Guid: guid,
	})

	if err != nil {
		return nil, err
	}
	if resp.Error != "" {
		return nil, errors.New("get resp.Error: " + resp.Error)
	}

	return resp.Data, nil
}

func (c *GRPCClient) Insert(ctx context.Context, section uint32, data GRPCData) (string, error) {
	resp, err := c.client.Insert(ctx, &grpcproto.InsertRequest{
		Section: section,
		Data:    data,
	})

	if err != nil {
		return "", err
	}
	if resp.Error != "" {
		return "", errors.New("insert resp.Error: " + resp.Error)
	}

	return resp.Guid, nil
}

func (c *GRPCClient) Update(ctx context.Context, guid string, data GRPCData) error {
	resp, err := c.client.Update(ctx, &grpcproto.UpdateRequest{
		Guid: guid,
		Data: data,
	})

	if err != nil {
		return err
	}
	if resp.Error != "" {
		return fmt.Errorf("update resp.Error: " + resp.Error)
	}

	return nil
}

func (c *GRPCClient) Remove(ctx context.Context, guid string) error {
	resp, err := c.client.Remove(ctx, &grpcproto.RemoveRequest{
		Guid: guid,
	})

	if err != nil {
		return err
	}
	if resp.Error != "" {
		return errors.New("remove resp.Error: " + resp.Error)
	}

	return nil
}
