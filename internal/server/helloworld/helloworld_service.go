package helloworld

import (
	"context"

	api "github.com/achew22/toy-project/api/v1"
)

// HelloWorldService implements the HelloWorldServer interface
type HelloWorldService struct {
	api.UnimplementedHelloWorldServer
}

// Greet implements the Greet method of the HelloWorldServer interface
func (s *HelloWorldService) Greet(ctx context.Context, req *api.GreetRequest) (*api.GreetResponse, error) {
	message := "Hello, " + req.GetName()
	return &api.GreetResponse{Message: message}, nil
}
