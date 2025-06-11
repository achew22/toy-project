// Package goldentest provides a framework for golden file testing with support for
// both single-file tests and multi-step sequential tests.
//
// # Overview
//
// Golden file testing compares the output of test functions against expected
// "golden" files. This package supports two testing modes:
//
//   - One-shot tests: Single input file produces single output
//   - Step tests: Sequential processing of numbered input files
//
// # Basic Usage
//
// Create a TestConfig with the appropriate test function and call RunTests:
//
//		config := &goldentest.TestConfig[*MyResult]{
//			InputExt:         ".hcl",
//			ErrorOutputExt:   ".txt",
//			SuccessOutputExt: ".json",
//	 	  TestOneShotFunc: func(filePath string, data []byte) (*MyResult, error) {
//	 	  	return processInput(data)
//	 	  },
//	 	  ErrorFunc: func(err error) []byte {
//	 	  	return []byte(err.Error())
//	 	  }
//		},
//		config.RunTests(t, "testdata")
//
// # One-Shot Tests
//
// For one-shot tests, set TestOneShotFunc. Each input file (e.g., "example.hcl")
// is processed independently, with output compared against "example.out.json"
// for success cases or "example.out.txt" for error cases (files starting
// with ErrorPrefix).
//
// # Step Tests
//
// For step tests, set StepTestFunc. Input files are numbered sequentially
// (1.hcl, 2.hcl, etc.) within subdirectories, and each step's output is
// compared against corresponding output files (1.out.json, 2.out.json, etc.).
//
//			config := &goldentest.TestConfig[*StepResult]{
//				InputExt:         ".in.textpb",
//				ErrorOutputExt:   ".txt",
//				SuccessOutputExt: ".textpb",
//				DiffOpts:         []cmp.Option{protocmp.Transform()},
//			  StepTestFunc: func(stepFile goldentest.StepFile) (*StepResult, error) {
//			  	return processStep(stepFile.Data)
//			  },
//			  ErrorFunc: func(err error) []byte {
//			  	return []byte(err.Error())
//			  }
//		 }
//	  config.RunTests(t, "testdata")
//
// # Configuration Rules
//
// TestConfig must have exactly one of TestOneShotFunc or StepTestFunc set:
//   - Setting both will cause RunTests to fail with t.Fatal
//   - Setting neither will cause RunTests to fail with t.Fatal
//
// Error handling configuration:
//   - ErrorFunc and ErrorOutputExt must both be set or both unset
//   - ErrorPrefix is optional and defaults to "error_" if ErrorFunc is set
//   - If ErrorFunc is not set, tests that return errors will fail immediately
//
// # File Organization
//
// One-shot tests expect files directly in the test directory:
//
//	testdata/
//	  ├── valid_input.hcl → valid_input.out.json
//	  ├── error_case.hcl → error_case.out.txt
//	  └── another.hcl → another.out.json
//
// Step tests expect subdirectories with numbered files:
//
//	testdata/
//	  ├── simple_flow/
//	  │   ├── 1.in.textpb → 1.out.textpb
//	  │   └── 2.in.textpb → 2.out.textpb
//	  └── error_scenario/
//	      ├── 1.in.textpb → 1.out.textpb
//	      └── 2.in.textpb → 2.out.txt (error)
//
// # Updating Golden Files
//
// Use the -update flag to regenerate expected output files:
//
//	go test -update ./path/to/tests
package goldentest

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"google.golang.org/protobuf/encoding/prototext"
	"google.golang.org/protobuf/proto"
)

// Update is a flag that controls whether golden files should be updated
var Update = flag.Bool("update", false, "update .out files if there is a difference")

// TestConfig holds configuration for golden file testing.
//
// TestConfig is generic over the result type T that test functions return.
// Exactly one of TestFunc or StepTestFunc must be set to determine the testing mode.
//
// Configuration fields:
//   - InputExt: File extension for input files (e.g., ".hcl", ".in.textpb")
//   - SuccessOutputExt: Extension for success output files (e.g., ".json", ".textpb")
//   - Formatter: Converts result values to bytes for golden file storage
//   - Loader: Converts bytes from golden files back to result values for comparison
//   - DiffOpts: Additional options for cmp.Diff (e.g., protocmp.Transform() for protobuf)
//
// Test function fields (exactly one must be set):
//   - TestOneShotFunc: For one-shot tests, processes individual input files
//   - StepTestFunc: For step tests, processes sequential numbered files
//
// Error handling fields (all three must be set together, or all left unset):
//   - ErrorFunc: Converts errors to byte representation for comparison
//   - ErrorPrefix: Prefix to identify error test case files (e.g., "error_")
//   - ErrorOutputExt: Extension for error output files (e.g., ".out.txt")
//
// The RunTests method automatically dispatches to the appropriate test mode
// based on which test function is configured.
type TestConfig[T any] struct {
	// InputExt is the file extension for input files (e.g., ".hcl", ".in.textpb")
	InputExt string
	// ErrorPrefix is the prefix used to identify error test cases (e.g., "error_").
	// If not specified but ErrorFunc is set, defaults to "error_".
	// Only valid if ErrorFunc is also set.
	ErrorPrefix string
	// ErrorOutputExt is the file extension for error output files (e.g., ".txt").
	// The framework automatically adds ".out" prefix, so ".txt" becomes ".out.txt".
	// Must be set together with ErrorFunc, or left empty if ErrorFunc is unset.
	ErrorOutputExt string
	// SuccessOutputExt is the file extension for success output files (e.g., ".json", ".textpb").
	// The framework automatically adds ".out" prefix, so ".json" becomes ".out.json".
	SuccessOutputExt string

	// Formatter converts result values to bytes for golden file storage.
	// If not set, DefaultFormatter[T]() will be used.
	Formatter Formatter[T]

	// Loader converts bytes from golden files back to result values for comparison.
	// If not set, DefaultLoader[T]() will be used.
	Loader Loader[T]

	// DiffOpts are additional options passed to cmp.Diff for comparing values.
	// For protobuf messages, typically include protocmp.Transform().
	// If not set, only cmpopts.EquateEmpty() will be used.
	DiffOpts []cmp.Option

	// TestOneShotFunc processes input data for one-shot tests. Set this for single-file golden tests.
	// Must not be set if StepTestFunc is set.
	TestOneShotFunc TestOneShotFunc[T]

	// StepTestFunc processes individual steps for multi-step tests. Set this for sequential golden tests.
	// Must not be set if TestOneShotFunc is set.
	StepTestFunc StepTestFunc[T]

	// ErrorFunc converts errors to byte representation for golden file comparison.
	// Must be set together with ErrorPrefix and ErrorOutputExt, or left nil with both unset.
	// If error handling is disabled, tests that return errors will fail immediately.
	ErrorFunc ErrorFunc
}

// TestOneShotFunc is a function that processes input data for one-shot golden tests.
//
// Used for single-file golden tests where each input file is processed independently.
// The function receives the full file path and content, and should return either a
// result of type T for successful processing, or an error for failed processing.
//
// Parameters:
//   - filePath: Full path to the input file being processed
//   - data: Raw content of the input file
//
// Returns:
//   - T: Result of processing the input (for success cases)
//   - error: Error encountered during processing (for error cases)
//
// Example:
//
//	config.TestOneShotFunc = func(filePath string, data []byte) (*Config, error) {
//		return parseConfig(filePath, data)
//	}
type TestOneShotFunc[T any] func(filePath string, data []byte) (T, error)

// ErrorFunc is a function that extracts error text from an error
type ErrorFunc func(err error) []byte

// Formatter is a function that converts a result of type T to bytes for golden file storage.
//
// Used to serialize test results into a consistent byte format for writing to golden files.
// The formatter should produce a stable, readable representation that can be reconstructed
// by the corresponding Loader.
//
// Parameters:
//   - value: The result value to format
//
// Returns:
//   - []byte: Serialized representation of the value
//   - error: Error if formatting fails
//
// Example:
//
//	config.Formatter = func(value *MyType) ([]byte, error) {
//		return json.MarshalIndent(value, "", "  ")
//	}
type Formatter[T any] func(value T) ([]byte, error)

// Loader is a function that converts bytes from a golden file back to type T.
//
// Used to deserialize golden file content back into the original type T for comparison.
// The loader should be symmetric to the Formatter - able to reconstruct values that
// were serialized by the formatter.
//
// Parameters:
//   - data: The serialized data from the golden file
//
// Returns:
//   - T: Deserialized value
//   - error: Error if loading fails
//
// Example:
//
//	config.Loader = func(data []byte) (*MyType, error) {
//		var result MyType
//		err := json.Unmarshal(data, &result)
//		return &result, err
//	}
type Loader[T any] func(data []byte) (T, error)

// DefaultFormatter returns a default formatter that handles common types.
//
// The default formatter uses the following strategy:
//   - string: Returns the string as bytes
//   - []byte: Returns the bytes directly
//   - proto.Message: Uses prototext marshaling with indentation
//   - Everything else: Uses JSON marshaling with indentation
//
// This covers the most common use cases and provides a reasonable default
// for most golden file testing scenarios.
func DefaultFormatter[T any]() Formatter[T] {
	return func(value T) ([]byte, error) {
		// Handle the most common cases first
		switch v := any(value).(type) {
		case string:
			return []byte(v), nil
		case []byte:
			return v, nil
		case proto.Message:
			// Use prototext for proto messages with nice formatting
			return prototext.MarshalOptions{
				Multiline: true,
				Indent:    "  ",
			}.Marshal(v)
		default:
			// Fall back to JSON for everything else
			return json.MarshalIndent(value, "", "  ")
		}
	}
}

// DefaultLoader returns a default loader that handles common types.
//
// The default loader uses the following strategy:
//   - string: Converts bytes to string
//   - []byte: Returns the bytes directly
//   - proto.Message: Uses prototext unmarshaling
//   - Everything else: Uses JSON unmarshaling
//
// This is symmetric to DefaultFormatter and handles the same type cases.
func DefaultLoader[T any]() Loader[T] {
	return func(data []byte) (T, error) {
		var zero T

		// Handle the most common cases first
		switch any(zero).(type) {
		case string:
			return any(string(data)).(T), nil
		case []byte:
			return any(data).(T), nil
		case proto.Message:
			// Create a new instance of the proto message type using ProtoReflect
			// Get the message descriptor from a zero value
			var zeroPtr *T = new(T)
			if msg, ok := any(*zeroPtr).(proto.Message); ok {
				newMsg := msg.ProtoReflect().New().Interface()
				err := prototext.Unmarshal(data, newMsg)
				return any(newMsg).(T), err
			}
			return zero, fmt.Errorf("type %T does not implement proto.Message", zero)
		default:
			// Fall back to JSON for everything else
			var result T
			err := json.Unmarshal(data, &result)
			return result, err
		}
	}
}

// RunTests runs golden file tests for all files in the specified directory.
//
// This is the main entry point for the golden test framework. It automatically
// determines the appropriate test mode based on the configured test functions:
//
//   - If TestOneShotFunc is set (and StepTestFunc is not): Runs one-shot tests
//   - If StepTestFunc is set (and TestOneShotFunc is not): Runs step tests
//   - If both are set: Calls t.Fatal with configuration error
//   - If neither are set: Calls t.Fatal with configuration error
//
// One-shot tests process each input file independently, comparing the result
// against corresponding golden output files in the same directory.
//
// Step tests process subdirectories containing numbered input files sequentially,
// comparing each step's result against numbered output files.
//
// Parameters:
//   - t: The testing.T instance for test execution and reporting
//   - dir: The directory containing test files (for one-shot) or subdirectories (for step tests)
//
// The method handles error cases (files prefixed with ErrorPrefix) by comparing
// the error output against .out.txt files, and success cases by comparing the
// result against output files using the configured Formatter and Loader.
func (config *TestConfig[T]) RunTests(t *testing.T, dir string) {
	// Check which test functions are set and dispatch accordingly
	oneShotFuncSet := config.TestOneShotFunc != nil
	stepTestFuncSet := config.StepTestFunc != nil

	if oneShotFuncSet && stepTestFuncSet {
		t.Fatal("TestConfig has both TestOneShotFunc and StepTestFunc set - only one should be configured")
	}
	if !oneShotFuncSet && !stepTestFuncSet {
		t.Fatal("TestConfig has neither TestOneShotFunc nor StepTestFunc set - one must be configured")
	}

	// Validate error handling configuration
	errorFuncSet := config.ErrorFunc != nil
	errorPrefixSet := config.ErrorPrefix != ""
	errorOutputExtSet := config.ErrorOutputExt != ""

	// ErrorFunc and ErrorOutputExt must both be set or both unset
	if (errorFuncSet && !errorOutputExtSet) || (!errorFuncSet && errorOutputExtSet) {
		t.Fatal("TestConfig error handling fields ErrorFunc and ErrorOutputExt must both be set or both unset")
	}

	// ErrorPrefix is only valid if ErrorFunc is set
	if errorPrefixSet && !errorFuncSet {
		t.Fatal("TestConfig ErrorPrefix is set but ErrorFunc is not - ErrorPrefix requires ErrorFunc to be set")
	}

	// Set default ErrorPrefix if error handling is enabled but no prefix specified
	if errorFuncSet && config.ErrorPrefix == "" {
		config.ErrorPrefix = "error_"
	}

	// Set default formatter and loader if not provided
	if config.Formatter == nil {
		config.Formatter = DefaultFormatter[T]()
	}
	if config.Loader == nil {
		config.Loader = DefaultLoader[T]()
	}

	if oneShotFuncSet {
		config.runOneShotTests(t, dir)
	} else {
		config.runStepTests(t, dir)
	}
}

// runOneShotTests runs golden file tests for all files in the specified directory
func (config *TestConfig[T]) runOneShotTests(t *testing.T, dir string) {
	files, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("failed to read testdata directory: %v", err)
	}

	for _, file := range files {
		if filepath.Ext(file.Name()) != config.InputExt {
			continue
		}

		t.Run(file.Name(), func(t *testing.T) {
			filePath := filepath.Join(dir, file.Name())
			data, err := os.ReadFile(filePath)
			if err != nil {
				t.Fatalf("failed to read file %s: %v", file.Name(), err)
			}

			outputFile := strings.TrimSuffix(file.Name(), config.InputExt)
			result, testErr := config.TestOneShotFunc(filePath, data)

			// Check if error handling is configured
			errorHandlingEnabled := config.ErrorFunc != nil

			if errorHandlingEnabled && strings.HasPrefix(file.Name(), config.ErrorPrefix) {
				// This is an error test case
				if testErr == nil {
					t.Errorf("expected error for file %s, but got none", file.Name())
					return
				}
				config.testErrorCase(t, dir, file.Name(), outputFile, testErr, config.ErrorFunc)
			} else {
				// This is a success test case (or error handling is disabled)
				if testErr != nil {
					if !errorHandlingEnabled {
						t.Errorf("test failed for file %s: %v", file.Name(), testErr)
						return
					}
					// Error handling is enabled but this isn't an error test case
					t.Errorf("unexpected error for file %s: %v", file.Name(), testErr)
					return
				}
				config.testSuccessCase(t, dir, file.Name(), outputFile, result, testErr)
			}
		})
	}
}

func (config *TestConfig[T]) testErrorCase(t *testing.T, dir, fileName, outputFile string, testErr error, errorFunc ErrorFunc) {
	outputFile += ".out" + config.ErrorOutputExt
	if testErr == nil {
		t.Errorf("expected error for file %s, but got none", fileName)
		return
	}

	expectedError, readErr := os.ReadFile(filepath.Join(dir, outputFile))
	if readErr != nil {
		t.Logf("failed to read expected error output file: %v", readErr)
	}

	actualError := errorFunc(testErr)
	if !bytes.Equal(expectedError, actualError) {
		if *Update {
			if writeErr := os.WriteFile(filepath.Join(dir, outputFile), actualError, 0644); writeErr != nil {
				t.Errorf("failed to update error output file: %v", writeErr)
			}
			return
		}
		t.Errorf("error output mismatch for file %s:\nExpected:\n%s\nGot:\n%s", fileName, expectedError, actualError)
	}
}

func (config *TestConfig[T]) testSuccessCase(t *testing.T, dir, fileName, outputFile string, result T, testErr error) {
	outputFile += ".out" + config.SuccessOutputExt
	if testErr != nil {
		t.Errorf("unexpected error for file %s: %v", fileName, testErr)
		return
	}

	// Use the configured formatter and loader (defaults set in RunTests)

	expectedData, readErr := os.ReadFile(filepath.Join(dir, outputFile))
	if readErr != nil {
		t.Logf("failed to read expected output file: %v", readErr)
	}

	// Load expected value from golden file
	expected, loadErr := config.Loader(expectedData)
	if loadErr != nil {
		t.Errorf("failed to load expected value from %s: %v", outputFile, loadErr)
		return
	}

	// Set up diff options
	var diffOpts []cmp.Option
	diffOpts = append(diffOpts, cmpopts.EquateEmpty())
	diffOpts = append(diffOpts, config.DiffOpts...)

	// Compare the actual T objects
	if diff := cmp.Diff(expected, result, diffOpts...); diff != "" {
		if *Update {
			// Format the actual result for writing to golden file
			actualData, formatErr := config.Formatter(result)
			if formatErr != nil {
				t.Errorf("failed to format result for %s: %v", fileName, formatErr)
				return
			}

			if writeErr := os.WriteFile(filepath.Join(dir, outputFile), actualData, 0644); writeErr != nil {
				t.Errorf("failed to update output file %s: %v", outputFile, writeErr)
			}
			return
		}
		t.Errorf("output mismatch for file %s (-expected +got):\n%s", fileName, diff)
	}
}
