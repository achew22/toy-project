// Package client provides a unified client interface for gRPC services.
package client

import (
	"context"
	"fmt"

	api "github.com/achew22/toy-project/api/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Request and Response types are generated from client.proto
//go:generate make protos

type Client struct {
	helloworldClient api.HelloWorldClient
}

func NewClient(conn grpc.ClientConnInterface) *Client {
	return &Client{
		helloworldClient: api.NewHelloWorldClient(conn),
	}
}

func (c *Client) Execute(ctx context.Context, req *Request) (*Response, error) {
	switch r := req.Request.(type) {
	case *Request_GreetRequest:
		resp, err := c.helloworldClient.Greet(ctx, r.GreetRequest)
		if err != nil {
			st, _ := status.FromError(err)
			return &Response{
				Response: &Response_Status{
					Status: st.Proto(),
				},
			}, nil
		}
		return &Response{
			Response: &Response_GreetResponse{
				GreetResponse: resp,
			},
		}, nil
	default:
		return &Response{
			Response: &Response_Status{
				Status: status.New(codes.Unimplemented, fmt.Sprintf("unimplemented request type: %T", r)).Proto(),
			},
		}, nil
	}
}
