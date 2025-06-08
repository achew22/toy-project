package server

import (
	"context"
	"log"
	"net"

	api "github.com/achew22/toy-project/api/v1"
	"github.com/achew22/toy-project/internal/server/helloworld"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type Server struct {
	grpcServer *grpc.Server
}

func NewServer() *Server {
	s := &Server{
		grpcServer: grpc.NewServer(),
	}
	s.register()
	return s
}

func (s *Server) register() {
	helloworldService := &helloworld.HelloWorldService{}
	api.RegisterHelloWorldServer(s.grpcServer, helloworldService)
	reflection.Register(s.grpcServer)
}

func (s *Server) Run(ctx context.Context, address string) error {
	lis, err := net.Listen("tcp", address)
	if err != nil {
		return err
	}
	return s.Serve(ctx, lis)
}

func (s *Server) Serve(ctx context.Context, lis net.Listener) error {
	go func() {
		<-ctx.Done()
		log.Println("Shutting down gRPC server...")
		s.grpcServer.GracefulStop()
	}()

	log.Printf("Starting gRPC server on %s\n", lis.Addr().String())
	if err := s.grpcServer.Serve(lis); err != nil {
		return err
	}

	return nil
}

func (s *Server) GRPCServer() *grpc.Server {
	return s.grpcServer
}

func (s *Server) Stop() {
	s.grpcServer.Stop()
}

func (s *Server) GracefulStop() {
	s.grpcServer.GracefulStop()
}
