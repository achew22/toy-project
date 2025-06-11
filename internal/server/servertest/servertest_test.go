package servertest

import (
	"context"
	"testing"
	"time"

	"google.golang.org/grpc/connectivity"
	"google.golang.org/grpc/reflection/grpc_reflection_v1alpha"
)

func TestServerTest_New(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	server := New(ctx)
	defer server.Close()

	if server.Address() == "" {
		t.Fatal("Expected non-empty address")
	}

	if server.Server() == nil {
		t.Fatal("Expected non-nil gRPC server")
	}

	if server.Listener() == nil {
		t.Fatal("Expected non-nil listener")
	}
}

func TestServerTest_NewClientConn(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	server := New(ctx)
	defer server.Close()

	connCtx, connCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer connCancel()

	conn, err := server.NewClientConn(connCtx)
	if err != nil {
		t.Fatalf("Failed to create client connection: %v", err)
	}
	defer conn.Close()

	// Wait for connection to be ready
	if !conn.WaitForStateChange(connCtx, connectivity.Connecting) {
		t.Fatal("Connection failed to establish")
	}

	// Test that we can use the reflection service
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
		t.Fatal("No services found")
	}
}

func TestServerTest_ContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	server := New(ctx)

	// Establish a connection first
	connCtx, connCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer connCancel()

	conn, err := server.NewClientConn(connCtx)
	if err != nil {
		t.Fatalf("Failed to create client connection: %v", err)
	}
	defer conn.Close()

	// Wait for connection to be ready
	if !conn.WaitForStateChange(connCtx, connectivity.Connecting) {
		t.Fatal("Connection failed to establish")
	}

	// Cancel the context
	cancel()

	// Wait for the connection state to change indicating server shutdown
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()

	for conn.GetState() == connectivity.Ready {
		if !conn.WaitForStateChange(shutdownCtx, conn.GetState()) {
			t.Fatal("Server did not shut down within timeout")
		}
	}

	if conn.GetState() == connectivity.Ready {
		t.Fatal("Expected connection to be closed after context cancellation")
	}
}
