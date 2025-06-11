# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Common Commands

- **Run tests**: `go test ./...` (all packages) or `go test ./internal/config` (specific package)
- **Run single test**: `go test -run TestFunctionName ./package/path`
- **Update test outputs**: `go test -update ./package/path` (for golden file tests)
- **Build**: `go build ./cmd/server`
- **Run server**: `go run ./cmd/server`
- **Generate protobuf**: `make protos` (uses buf.gen.yaml configuration)
- **Lint protobuf**: `make lint`

## Architecture Overview

This is a Go-based gRPC server project with HCL configuration parsing:

### Core Components

- **cmd/server/main.go**: Entry point with signal handling and error propagation using custom ExitError type
- **internal/server/**: gRPC server implementation with graceful shutdown and reflection service
- **internal/config/**: HCL configuration parser with validation for server settings
- **api/v1/**: Protocol buffer definitions for HelloWorld service
- **internal/server/helloworld/**: HelloWorld gRPC service implementation

### Key Patterns

- **Configuration**: Uses HashiCorp HCL format with custom function support and detailed error diagnostics
- **Testing**: Golden file pattern for config tests (testdata/*.hcl with corresponding *.out.json/.out.txt files)
- **gRPC**: Server includes reflection service for development tooling
- **Error Handling**: Custom ExitError type for proper exit codes from main function

### Protocol Buffers

- Uses buf for protobuf management with local tool paths (internal/tools/)
- Code generation configured in buf.gen.yaml with source_relative paths
- Always use `make protos` to generate protos (protobufs)

### Configuration Format

Server expects HCL files with required `server` block containing `listening_address` in host:port format. Supports custom HCL functions for dynamic configuration.

### General instructions

- Before considering any work done, be sure to run `make fmt` to ensure the files are formatted properly.
