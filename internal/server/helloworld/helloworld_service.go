package helloworld

import (
	"context"
	apiv1 "path/to/your/api/v1" // Update this import path to match your project structure
)

// HelloWorldService implements the HelloWorldServer interface
type HelloWorldService struct {
	apiv1.UnimplementedHelloWorldServer
}

// Greet implements the Greet method of the HelloWorldServer interface
func (s *HelloWorldService) Greet(ctx context.Context, req *apiv1.GreetRequest) (*apiv1.GreetResponse, error) {
	message := "Hello, " + req.GetName()
	return &apiv1.GreetResponse{Message: message}, nil
}
