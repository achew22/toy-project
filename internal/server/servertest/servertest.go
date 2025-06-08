package servertest

import (
	"context"
	"net"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/achew22/toy-project/internal/server"
)

// ServerTest represents a test gRPC server for testing purposes.
type ServerTest struct {
	server   *server.Server
	listener net.Listener
	address  string
	ctx      context.Context
	cancel   context.CancelFunc
}

// New creates a new test gRPC server listening on a loopback address.
// The server's lifecycle is tied to the provided context.
// It returns a ServerTest that can be used for testing gRPC services.
func New(ctx context.Context) *ServerTest {
	lis, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		panic(err)
	}

	srv := server.NewServer()

	serverCtx, cancel := context.WithCancel(ctx)

	s := &ServerTest{
		server:   srv,
		listener: lis,
		address:  lis.Addr().String(),
		ctx:      serverCtx,
		cancel:   cancel,
	}

	go func() {
		if err := srv.Serve(serverCtx, lis); err != nil {
			// Server was closed, ignore the error
		}
	}()

	return s
}

// Close shuts down the test server and releases its resources.
func (s *ServerTest) Close() {
	s.cancel()
	s.server.Stop()
	s.listener.Close()
}

// GracefulStop gracefully stops the test server.
func (s *ServerTest) GracefulStop() {
	s.cancel()
	s.server.GracefulStop()
	s.listener.Close()
}

// Server returns the underlying gRPC server for registering services.
func (s *ServerTest) Server() *grpc.Server {
	return s.server.GRPCServer()
}

// Address returns the server's listening address.
func (s *ServerTest) Address() string {
	return s.address
}

// Listener returns the underlying net.Listener.
func (s *ServerTest) Listener() net.Listener {
	return s.listener
}

// NewClientConn creates a new gRPC client connection to the test server.
// The caller is responsible for closing the connection.
func (s *ServerTest) NewClientConn(ctx context.Context) (*grpc.ClientConn, error) {
	return grpc.DialContext(ctx, s.address, grpc.WithTransportCredentials(insecure.NewCredentials()))
}

// URL returns the server address in a format suitable for gRPC dial.
func (s *ServerTest) URL() string {
	return s.address
}
