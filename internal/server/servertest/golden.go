package servertest

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/encoding/prototext"
	"google.golang.org/protobuf/testing/protocmp"

	"github.com/achew22/toy-project/internal/goldentest"
	"github.com/achew22/toy-project/internal/server/servertest/client"

	pb "github.com/achew22/toy-project/internal/server/servertest/proto/v1"
)

// ServerFixture holds the server and client resources for testing
type ServerFixture struct {
	Server *ServerTest
	Client *client.Client
	Conn   *grpc.ClientConn
}

// RunGoldenStepTests runs golden step tests for gRPC server interactions.
// It starts a server once and reuses it across all test steps.
// Each step consists of a TestStepIn input and produces a TestStepOut output.
func RunGoldenStepTests(t *testing.T) {
	config := &goldentest.TestConfig[*pb.TestStepOut, *ServerFixture]{
		InputExt:         ".in.textpb",
		ErrorOutputExt:   ".txt",
		SuccessOutputExt: ".textpb",
		DiffOpts:         []cmp.Option{protocmp.Transform()},
		SetUp: func(t *testing.T) (*ServerFixture, error) {
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
			return &ServerFixture{
				Server: server,
				Client: grpcClient,
				Conn:   conn,
			}, nil
		},
		TearDown: func(t *testing.T, fixture *ServerFixture) error {
			fixture.Conn.Close()
			fixture.Server.Close()
			return nil
		},
		StepTestFunc: func(ctx context.Context, fixture *ServerFixture, stepFile goldentest.StepFile) (*pb.TestStepOut, error) {
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

		ErrorFunc: func(err error) []byte {
			return []byte(err.Error())
		},
	}

	config.RunTests(t, "testdata")
}
