package servertest

import (
	"context"
	"testing"

	"google.golang.org/protobuf/encoding/prototext"

	"github.com/achew22/toy-project/internal/golden"
	"github.com/achew22/toy-project/internal/server/servertest/client"

	pb "github.com/achew22/toy-project/internal/server/servertest/proto/v1"
)

// RunGoldenStepTests runs golden step tests for gRPC server interactions.
// It starts a server once and reuses it across all test steps.
// Each step consists of a TestStepIn input and produces a TestStepOut output.
func RunGoldenStepTests(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start the server once for all test steps
	server := New(ctx)
	defer server.Close()

	// Create client connection
	conn, err := server.NewClientConn(context.Background())
	if err != nil {
		t.Fatalf("Failed to create client connection: %v", err)
	}
	defer conn.Close()

	// Create the unified client
	grpcClient := client.NewClient(conn)

	config := &golden.TestConfig{
		TestDataDir:      "testdata",
		InputExt:         ".in.textpb",
		ErrorPrefix:      "error_",
		ErrorOutputExt:   ".out.txt",
		SuccessOutputExt: ".out.textpb",
		UsePrototext:     true,
	}

	stepTestFunc := func(stepFile golden.StepFile) (*pb.TestStepOut, error) {
		// Parse the input step
		stepIn := &pb.TestStepIn{}
		if err := prototext.Unmarshal(stepFile.Data, stepIn); err != nil {
			return nil, err
		}

		// Execute the RPC
		response, err := grpcClient.Execute(context.Background(), stepIn.Rpc)
		if err != nil {
			return nil, err
		}

		// Create the output step
		stepOut := &pb.TestStepOut{
			Rpc: response,
		}
		return stepOut, nil
	}

	errorFunc := func(err error) []byte {
		return []byte(err.Error())
	}

	golden.RunStepTests(t, config, stepTestFunc, errorFunc)
}
