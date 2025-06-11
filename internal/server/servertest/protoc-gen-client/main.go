package main

import (
	"fmt"
	"strings"

	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/types/pluginpb"
)

func main() {
	protogen.Options{}.Run(func(gen *protogen.Plugin) error {
		gen.SupportedFeatures = uint64(pluginpb.CodeGeneratorResponse_FEATURE_PROTO3_OPTIONAL)

		// Find all services and their methods
		var services []*protogen.Service
		for _, file := range gen.Files {
			if !file.Generate {
				continue
			}
			services = append(services, file.Services...)
		}

		if len(services) == 0 {
			return nil
		}

		// Generate the client.proto file
		protoFile := gen.NewGeneratedFile("client.proto", "")
		generateProtoFile(protoFile, services)

		// Generate the client.go file
		goFile := gen.NewGeneratedFile("client.go", "github.com/achew22/toy-project/internal/server/servertest/client")
		generateGoFile(goFile, services)

		return nil
	})
}

func generateProtoFile(g *protogen.GeneratedFile, services []*protogen.Service) {
	g.P(`syntax = "proto3";`)
	g.P()
	g.P(`package cmd.achew.toyproject.api.v1;`)
	g.P()
	g.P(`option go_package = "github.com/achew22/toy-project/internal/server/servertest/client;client";`)
	g.P()
	g.P(`import "google/rpc/status.proto";`)

	// Import all the proto files that contain the method request/response types
	importedFiles := make(map[string]bool)
	for _, service := range services {
		file := service.Location.SourceFile
		if file != "" && !importedFiles[file] {
			// Extract relative path for import
			relativePath := strings.TrimPrefix(file, "api/v1/")
			if relativePath != file && relativePath != "client.proto" {
				g.P(`import "api/v1/` + relativePath + `";`)
				importedFiles[file] = true
			}
		}
	}
	g.P()

	// Generate Request message with oneof for each method
	g.P(`message Request {`)
	g.P(`  oneof request {`)
	fieldNum := 1
	for _, service := range services {
		for _, method := range service.Methods {
			methodName := strings.ToLower(method.GoName)
			typeName := string(method.Input.Desc.Name())
			g.P(fmt.Sprintf(`    %s %s_request = %d;`, typeName, methodName, fieldNum))
			fieldNum++
		}
	}
	g.P(`  }`)
	g.P(`}`)
	g.P()

	// Generate Response message with oneof for each method plus status
	g.P(`message Response {`)
	g.P(`  oneof response {`)
	g.P(`    google.rpc.Status status = 1;`)
	fieldNum = 2
	for _, service := range services {
		for _, method := range service.Methods {
			methodName := strings.ToLower(method.GoName)
			typeName := string(method.Output.Desc.Name())
			g.P(fmt.Sprintf(`    %s %s_response = %d;`, typeName, methodName, fieldNum))
			fieldNum++
		}
	}
	g.P(`  }`)
	g.P(`}`)
}

func generateGoFile(g *protogen.GeneratedFile, services []*protogen.Service) {
	g.P(`// Package client provides a unified client interface for gRPC services.`)
	g.P(`package client`)
	g.P()
	g.P(`import (`)
	g.P(`	"context"`)
	g.P(`	"fmt"`)
	g.P()
	g.P(`	api "github.com/achew22/toy-project/api/v1"`)
	g.P(`	"google.golang.org/grpc"`)
	g.P(`	"google.golang.org/grpc/codes"`)
	g.P(`	"google.golang.org/grpc/status"`)
	g.P(`)`)
	g.P()
	g.P(`// Request and Response types are generated from client.proto`)
	g.P(`//go:generate make protos`)
	g.P()

	// Generate Client struct
	g.P(`type Client struct {`)
	for _, service := range services {
		clientName := strings.ToLower(service.GoName) + "Client"
		g.P(fmt.Sprintf(`	%s api.%sClient`, clientName, service.GoName))
	}
	g.P(`}`)
	g.P()

	// Generate NewClient constructor
	g.P(`func NewClient(conn grpc.ClientConnInterface) *Client {`)
	g.P(`	return &Client{`)
	for _, service := range services {
		clientName := strings.ToLower(service.GoName) + "Client"
		g.P(fmt.Sprintf(`		%s: api.New%sClient(conn),`, clientName, service.GoName))
	}
	g.P(`	}`)
	g.P(`}`)
	g.P()

	// Generate Execute method
	g.P(`func (c *Client) Execute(ctx context.Context, req *Request) (*Response, error) {`)
	g.P(`	switch r := req.Request.(type) {`)

	for _, service := range services {
		for _, method := range service.Methods {
			methodName := strings.ToLower(method.GoName)
			clientName := strings.ToLower(service.GoName) + "Client"

			g.P(fmt.Sprintf(`	case *Request_%sRequest:`, strings.Title(methodName)))
			g.P(fmt.Sprintf(`		resp, err := c.%s.%s(ctx, r.%sRequest)`, clientName, method.GoName, strings.Title(methodName)))
			g.P(`		if err != nil {`)
			g.P(`			st, _ := status.FromError(err)`)
			g.P(`			return &Response{`)
			g.P(`				Response: &Response_Status{`)
			g.P(`					Status: st.Proto(),`)
			g.P(`				},`)
			g.P(`			}, nil`)
			g.P(`		}`)
			g.P(`		return &Response{`)
			g.P(fmt.Sprintf(`			Response: &Response_%sResponse{`, strings.Title(methodName)))
			g.P(fmt.Sprintf(`				%sResponse: resp,`, strings.Title(methodName)))
			g.P(`			},`)
			g.P(`		}, nil`)
		}
	}

	g.P(`	default:`)
	g.P(`		return &Response{`)
	g.P(`			Response: &Response_Status{`)
	g.P(`				Status: status.New(codes.Unimplemented, fmt.Sprintf("unimplemented request type: %T", r)).Proto(),`)
	g.P(`			},`)
	g.P(`		}, nil`)
	g.P(`	}`)
	g.P(`}`)
}
