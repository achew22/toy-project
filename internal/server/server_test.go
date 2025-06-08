package server

import (
	"context"
	"testing"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection/grpc_reflection_v1alpha"
)

func TestServer_Run(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	s := NewServer()
	address := "localhost:50051"

	go func() {
		if err := s.Run(ctx, address); err != nil {
			t.Errorf("Failed to run server: %v", err)
		}
	}()

	// Give the server a moment to start
	time.Sleep(1 * time.Second)

	conn, err := grpc.Dial(address, grpc.WithInsecure(), grpc.WithBlock(), grpc.WithTimeout(5*time.Second))
	if err != nil {
		t.Fatalf("Failed to connect to server: %v", err)
	}
	defer conn.Close()

	client := grpc_reflection_v1alpha.NewServerReflectionClient(conn)
	stream, err := client.ServerReflectionInfo(context.Background())
	if err != nil {
		t.Fatalf("Failed to create reflection client: %v", err)
	}

	if err := stream.Send(&grpc_reflection_v1alpha.ServerReflectionRequest{
		MessageRequest: &grpc_reflection_v1alpha.ServerReflectionRequest_ListServices{
			ListServices: "*",
		},
	}); err != nil {
		t.Fatalf("Failed to send reflection request: %v", err)
	}

	resp, err := stream.Recv()
	if err != nil {
		t.Fatalf("Failed to receive reflection response: %v", err)
	}

	if len(resp.GetListServicesResponse().Service) == 0 {
		t.Fatalf("No services found")
	}
}
