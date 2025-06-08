package client

import (
	"context"
	"testing"

	api "github.com/achew22/toy-project/api/v1"
	"github.com/achew22/toy-project/internal/server/servertest"
)

func TestClient_Execute(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Create a test server
	server := servertest.New(ctx)
	defer server.Close()

	// Create a client connection
	conn, err := server.NewClientConn(context.Background())
	if err != nil {
		t.Fatalf("Failed to create client connection: %v", err)
	}
	defer conn.Close()

	// Create the unified client
	client := NewClient(conn)

	// Test successful request
	req := &Request{
		Request: &Request_GreetRequest{
			GreetRequest: &api.GreetRequest{
				Name: "World",
			},
		},
	}

	resp, err := client.Execute(context.Background(), req)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	// Check that we got a successful response
	greetResp, ok := resp.Response.(*Response_GreetResponse)
	if !ok {
		t.Fatalf("Expected GreetResponse, got %T", resp.Response)
	}

	if greetResp.GreetResponse.Message != "Hello, World" {
		t.Errorf("Expected 'Hello, World', got %q", greetResp.GreetResponse.Message)
	}
}