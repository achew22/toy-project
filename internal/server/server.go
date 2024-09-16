package server

import (
	"context"
	"log"
	"net"

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
	reflection.Register(s.grpcServer)
	return s
}

func (s *Server) Run(ctx context.Context, address string) error {
	lis, err := net.Listen("tcp", address)
	if err != nil {
		return err
	}

	go func() {
		<-ctx.Done()
		log.Println("Shutting down gRPC server...")
		s.grpcServer.GracefulStop()
	}()

	log.Printf("Starting gRPC server on %s\n", address)
	if err := s.grpcServer.Serve(lis); err != nil {
		return err
	}

	return nil
}
