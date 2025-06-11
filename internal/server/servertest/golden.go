package servertest

import (
	"context"
	"testing"

	"google.golang.org/grpc"
	"google.golang.org/protobuf/encoding/prototext"

	"github.com/achew22/toy-project/internal/goldentest"
	"github.com/achew22/toy-project/internal/server/servertest/client"

	pb "github.com/achew22/toy-project/internal/server/servertest/proto/v1"
)

// serverFixture holds the server and client resources for testing
type serverFixture struct {
	Server *ServerTest
	Client *client.Client
	Conn   *grpc.ClientConn
}

var testSuite = goldentest.NewStepConfig(
	func(ctx context.Context, fixture *serverFixture, stepFile goldentest.StepFile) (*pb.TestStepOut, error) {
		// Parse the input step
		stepIn := &pb.TestStepIn{}
		if err := prototext.Unmarshal(stepFile.Data, stepIn); err != nil {
			return nil, err
		}

		// Execute the RPC
		response, err := fixture.Client.Execute(ctx, stepIn.Rpc)
		if err != nil {
			return nil, err
		}

		// Create the output step
		stepOut := &pb.TestStepOut{
			Rpc: response,
		}
		return stepOut, nil
	},
).
	WithInputExt(".textpb").
	WithSetUp(func(t *testing.T) (*serverFixture, error) {
		// Start the server once for all test steps
		server := New(t.Context())

		// Create client connection
		conn, err := server.NewClientConn(context.Background())
		if err != nil {
			server.Close()
			return nil, err
		}

		// Create the unified client
		grpcClient := client.NewClient(conn)
		return &serverFixture{
			Server: server,
			Client: grpcClient,
			Conn:   conn,
		}, nil
	}).
	WithTearDown(func(t *testing.T, fixture *serverFixture) error {
		fixture.Conn.Close()
		fixture.Server.Close()
		return nil
	}).
	Build()

// RunGoldenStepTests runs golden step tests for gRPC server interactions.
// It starts a server once and reuses it across all test steps.
// Each step consists of a TestStepIn input and produces a TestStepOut output.
func RunGoldenStepTests(t *testing.T) {
	testSuite.RunTests(t, "testdata")
}
