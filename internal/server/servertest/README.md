# ServerTest Golden Step Testing Framework

This package provides a golden file testing framework for gRPC services that allows you to test complete server interactions using human-readable protobuf text files.

## Overview

The framework starts a single gRPC server and reuses it across multiple test steps. Each test step consists of:
- **Input**: A `.in.textpb` file containing a `TestStepIn` message with the RPC request
- **Output**: A `.out.textpb` file containing the expected `TestStepOut` message with the RPC response

## Quick Start

### 1. Create a Test Function

In your service package, create a test file using an external test package to avoid import cycles:

```go
package myservice_test

import (
    "testing"
    
    "github.com/achew22/toy-project/internal/server/servertest"
)

func TestMyService_Golden(t *testing.T) {
    servertest.RunGoldenStepTests(t)
}
```

### 2. Create Test Data Directory Structure

Create a `testdata/` directory in your service package with test case subdirectories:

```
internal/server/myservice/
├── myservice_test.go
└── testdata/
    ├── simple_case/
    │   └── 1.in.textpb
    ├── multi_step_scenario/
    │   ├── 1.in.textpb
    │   ├── 2.in.textpb
    │   └── 3.in.textpb
    └── error_case/
        └── 1.in.textpb
```

### 3. Write Input Files

Create `.in.textpb` files with `TestStepIn` messages. Each file must be numbered sequentially starting from 1:

**`testdata/simple_case/1.in.textpb`**:
```textpb
actor: "user123"
rpc: {
  greet_request: {
    name: "World"
  }
}
```

**`testdata/multi_step_scenario/1.in.textpb`**:
```textpb
actor: "alice"
rpc: {
  greet_request: {
    name: "Alice"
  }
}
```

**`testdata/multi_step_scenario/2.in.textpb`**:
```textpb
actor: "bob"
rpc: {
  greet_request: {
    name: "Bob"
  }
}
```

### 4. Generate Golden Files

Run your test with the `-update` flag to automatically generate the expected output files:

```bash
go test ./internal/server/myservice -update -v
```

This will create corresponding `.out.textpb` files:

**`testdata/simple_case/1.out.textpb`**:
```textpb
rpc: {
  greet_response: {
    message: "Hello, World"
  }
}
```

### 5. Run Tests

Run your tests normally to verify against the golden files:

```bash
go test ./internal/server/myservice -v
```

## File Naming Conventions

- **Input files**: `{step_number}.in.textpb` (e.g., `1.in.textpb`, `2.in.textpb`)
- **Output files**: `{step_number}.out.textpb` (e.g., `1.out.textpb`, `2.out.textpb`)
- **Error files**: `{step_number}.out.txt` (for error test cases)

## Test Structure

### TestStepIn Message

Each input file contains a `TestStepIn` message with:
- `actor`: String identifying who is making the request
- `rpc`: The actual gRPC request message (supports any service method)

### TestStepOut Message

Each output file contains a `TestStepOut` message with:
- `rpc`: The expected gRPC response message or error status

## Test Case Types

### Success Cases

Regular test case directories contain sequential steps that should all succeed:

```
testdata/user_workflow/
├── 1.in.textpb    # First step
├── 1.out.textpb   # Expected response
├── 2.in.textpb    # Second step
└── 2.out.textpb   # Expected response
```

### Error Cases

Test case directories prefixed with `error_` test error conditions:

```
testdata/error_invalid_input/
├── 1.in.textpb    # Input that should cause an error
└── 1.out.txt      # Expected error message
```

## Step Execution Model

1. **Server Lifecycle**: One server is started and reused for all test cases
2. **Sequential Execution**: Steps within a test case are executed in order (1, 2, 3...)
3. **Independent Verification**: Each step's response is verified against its paired output file
4. **State Persistence**: Server state persists between steps within a test case

## Advanced Features

### Multi-Step Scenarios

Test complex workflows by creating multiple sequential steps:

```textpb
# 1.in.textpb - Create user
actor: "admin"
rpc: {
  create_user_request: {
    name: "john"
    email: "john@example.com"
  }
}

# 2.in.textpb - Greet user
actor: "john"
rpc: {
  greet_request: {
    name: "john"
  }
}
```

### Different RPC Methods

The framework supports any gRPC method defined in your service:

```textpb
# For HelloWorld.Greet
rpc: {
  greet_request: {
    name: "Alice"
  }
}

# For UserService.CreateUser
rpc: {
  create_user_request: {
    name: "Bob"
    email: "bob@example.com"
  }
}
```

### Error Testing

Test error conditions by prefixing directory names with `error_`:

```
testdata/error_empty_name/
├── 1.in.textpb    # Request with empty name
└── 1.out.txt      # Expected error text
```

## Best Practices

1. **Use descriptive directory names** that clearly indicate what scenario is being tested
2. **Keep steps focused** - each step should test one logical operation
3. **Use meaningful actor names** to make test scenarios more readable
4. **Start simple** with single-step tests before creating complex multi-step scenarios
5. **Review generated golden files** to ensure they contain expected values
6. **Version control golden files** to track changes in API behavior

## Troubleshooting

### Import Cycles

If you get import cycle errors, use an external test package:

```go
package myservice_test  // Note: _test suffix

import (
    "testing"
    "github.com/achew22/toy-project/internal/server/servertest"
)
```

### Missing Output Files

If tests fail with "file does not exist" errors, run with `-update` to generate golden files:

```bash
go test ./internal/server/myservice -update
```

### Step Validation Errors

Ensure step files are numbered sequentially starting from 1 with no gaps:
- ✅ `1.in.textpb`, `2.in.textpb`, `3.in.textpb`
- ❌ `1.in.textpb`, `3.in.textpb` (missing step 2)
- ❌ `0.in.textpb` (steps must start from 1)

### Protobuf Format Errors

Ensure your `.textpb` files use valid protobuf text format:
- Use proper message structure with `{}`
- Quote string values
- Use correct field names as defined in your `.proto` files

## Example

See `internal/server/helloworld/helloworld_service_test.go` for a complete working example.