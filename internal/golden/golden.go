package golden

import (
	"bytes"
	"encoding/json"
	"flag"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

// Update is a flag that controls whether golden files should be updated
var Update = flag.Bool("update", false, "update .out files if there is a difference")

// TestConfig holds configuration for golden file testing
type TestConfig struct {
	// TestDataDir is the directory containing test input and expected output files
	TestDataDir string
	// InputExt is the file extension for input files (e.g., ".hcl")
	InputExt string
	// ErrorPrefix is the prefix used to identify error test cases
	ErrorPrefix string
	// ErrorOutputExt is the file extension for error output files (e.g., ".out.txt")
	ErrorOutputExt string
	// SuccessOutputExt is the file extension for success output files (e.g., ".out.json")
	SuccessOutputExt string
}

// DefaultConfig returns a default TestConfig for HCL-based tests
func DefaultConfig() *TestConfig {
	return &TestConfig{
		TestDataDir:      "testdata",
		InputExt:         ".hcl",
		ErrorPrefix:      "error_",
		ErrorOutputExt:   ".out.txt",
		SuccessOutputExt: ".out.json",
	}
}

// TestFunc is a function that processes input data and returns either a result or an error
type TestFunc[T any] func(filePath string, data []byte) (T, error)

// ErrorFunc is a function that extracts error text from an error
type ErrorFunc func(err error) []byte

// RunTests runs golden file tests for all input files in the test data directory
func RunTests[T any](t *testing.T, config *TestConfig, testFunc TestFunc[T], errorFunc ErrorFunc) {
	files, err := os.ReadDir(config.TestDataDir)
	if err != nil {
		t.Fatalf("failed to read testdata directory: %v", err)
	}

	for _, file := range files {
		if filepath.Ext(file.Name()) != config.InputExt {
			continue
		}

		t.Run(file.Name(), func(t *testing.T) {
			filePath := filepath.Join(config.TestDataDir, file.Name())
			data, err := os.ReadFile(filePath)
			if err != nil {
				t.Fatalf("failed to read file %s: %v", file.Name(), err)
			}

			outputFile := strings.TrimSuffix(file.Name(), config.InputExt)
			result, testErr := testFunc(filePath, data)

			if strings.HasPrefix(file.Name(), config.ErrorPrefix) {
				// This is an error test case
				testErrorCase[T](t, config, file.Name(), outputFile, testErr, errorFunc)
			} else {
				// This is a success test case
				testSuccessCase[T](t, config, file.Name(), outputFile, result, testErr)
			}
		})
	}
}

func testErrorCase[T any](t *testing.T, config *TestConfig, fileName, outputFile string, testErr error, errorFunc ErrorFunc) {
	outputFile += config.ErrorOutputExt
	if testErr == nil {
		t.Errorf("expected error for file %s, but got none", fileName)
		return
	}

	expectedError, readErr := os.ReadFile(filepath.Join(config.TestDataDir, outputFile))
	if readErr != nil {
		t.Logf("failed to read expected error output file: %v", readErr)
	}

	actualError := errorFunc(testErr)
	if !bytes.Equal(expectedError, actualError) {
		if *Update {
			if writeErr := os.WriteFile(filepath.Join(config.TestDataDir, outputFile), actualError, 0644); writeErr != nil {
				t.Errorf("failed to update error output file: %v", writeErr)
			}
			return
		}
		t.Errorf("error output mismatch for file %s:\nExpected:\n%s\nGot:\n%s", fileName, expectedError, actualError)
	}
}

func testSuccessCase[T any](t *testing.T, config *TestConfig, fileName, outputFile string, result T, testErr error) {
	outputFile += config.SuccessOutputExt
	if testErr != nil {
		t.Errorf("unexpected error for file %s: %v", fileName, testErr)
		return
	}

	expectedData, readErr := os.ReadFile(filepath.Join(config.TestDataDir, outputFile))
	if readErr != nil {
		t.Logf("failed to read expected JSON output file: %v", readErr)
	}

	var expected T
	if err := json.Unmarshal(expectedData, &expected); err != nil {
		t.Errorf("failed to unmarshal expected JSON: %v", err)
		return
	}

	if diff := cmp.Diff(expected, result, cmpopts.EquateEmpty()); diff != "" {
		if *Update {
			actualData, marshalErr := json.MarshalIndent(result, "", "  ")
			if marshalErr != nil {
				t.Errorf("failed to marshal result to JSON: %v", marshalErr)
				return
			}
			if writeErr := os.WriteFile(filepath.Join(config.TestDataDir, outputFile), actualData, 0644); writeErr != nil {
				t.Errorf("failed to update JSON output file: %v", writeErr)
			}
			return
		}
		t.Errorf("output mismatch for file %s (-expected +got):\n%s", fileName, diff)
	}
}