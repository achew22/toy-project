package goldentest

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"google.golang.org/protobuf/encoding/prototext"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/testing/protocmp"
)

// StepTestFunc is a function that processes a single input file and returns either a result or an error
type StepTestFunc[T any] func(stepFile StepFile) (T, error)

// StepFile represents a single step in a sequence with its file path and data
type StepFile struct {
	// Step is the 1-based step number
	Step int
	// FilePath is the full path to the step file
	FilePath string
	// Data is the content of the step file
	Data []byte
}

// RunStepTests runs golden file tests in step mode for all directories in the test data directory
func RunStepTests[T any](t *testing.T, config *TestConfig, stepTestFunc StepTestFunc[T], errorFunc ErrorFunc) {
	entries, err := os.ReadDir(config.TestDataDir)
	if err != nil {
		t.Fatalf("failed to read testdata directory: %v", err)
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		t.Run(entry.Name(), func(t *testing.T) {
			stepDir := filepath.Join(config.TestDataDir, entry.Name())
			stepFiles, validateErr := validateAndLoadStepFiles(stepDir, config.InputExt, config)
			if validateErr != nil {
				t.Fatalf("failed to validate step directory %s: %v", entry.Name(), validateErr)
			}

			var results []T
			var testErr error

			// Execute stepTestFunc for each step file
			for _, stepFile := range stepFiles {
				result, err := stepTestFunc(stepFile)
				if err != nil {
					testErr = err
					break
				}
				results = append(results, result)
			}

			if strings.HasPrefix(entry.Name(), config.ErrorPrefix) {
				// This is an error test case - only test the final result
				if testErr == nil {
					t.Errorf("expected error for test %s, but got none", entry.Name())
					return
				}
				// Test error with final step number as filename
				finalStepNum := len(stepFiles)
				errorFile := fmt.Sprintf("%d%s", finalStepNum, config.ErrorOutputExt)
				testErrorCaseStep(t, config, stepDir, errorFile, testErr, errorFunc)
			} else {
				// This is a success test case - test each step result
				if testErr != nil {
					t.Errorf("unexpected error for test %s: %v", entry.Name(), testErr)
					return
				}
				testSuccessCaseSteps(t, config, stepDir, stepFiles, results)
			}
		})
	}
}

func testErrorCaseStep(t *testing.T, config *TestConfig, stepDir, errorFile string, testErr error, errorFunc ErrorFunc) {
	expectedError, readErr := os.ReadFile(filepath.Join(stepDir, errorFile))
	if readErr != nil {
		t.Logf("failed to read expected error output file: %v", readErr)
	}

	actualError := errorFunc(testErr)
	if !bytes.Equal(expectedError, actualError) {
		if *Update {
			if writeErr := os.WriteFile(filepath.Join(stepDir, errorFile), actualError, 0644); writeErr != nil {
				t.Errorf("failed to update error output file: %v", writeErr)
			}
			return
		}
		t.Errorf("error output mismatch for file %s:\nExpected:\n%s\nGot:\n%s", errorFile, expectedError, actualError)
	}
}

func testSuccessCaseSteps[T any](t *testing.T, config *TestConfig, stepDir string, stepFiles []StepFile, results []T) {
	if len(results) != len(stepFiles) {
		t.Errorf("expected %d results, got %d", len(stepFiles), len(results))
		return
	}

	var diffOpts []cmp.Option
	diffOpts = append(diffOpts, cmpopts.EquateEmpty())
	
	// If using prototext, use protocmp.Transform for proper protobuf comparison
	if config.UsePrototext {
		diffOpts = append(diffOpts, protocmp.Transform())
	}

	for i, result := range results {
		stepNum := stepFiles[i].Step
		outputFile := fmt.Sprintf("%d%s", stepNum, config.SuccessOutputExt)
		outputPath := filepath.Join(stepDir, outputFile)

		expectedData, readErr := os.ReadFile(outputPath)
		if readErr != nil {
			if config.UsePrototext {
				t.Logf("failed to read expected prototext output file %s: %v", outputFile, readErr)
			} else {
				t.Logf("failed to read expected JSON output file %s: %v", outputFile, readErr)
			}
		}

		var expected T
		var unmarshalErr error
		
		if config.UsePrototext {
			// For prototext, T must be a proto.Message
			if resultMsg, ok := any(result).(proto.Message); ok {
				expectedMsg := proto.Clone(resultMsg)
				proto.Reset(expectedMsg)
				unmarshalErr = prototext.Unmarshal(expectedData, expectedMsg)
				expected = any(expectedMsg).(T)
			} else {
				t.Errorf("result type %T does not implement proto.Message", result)
				return
			}
		} else {
			unmarshalErr = json.Unmarshal(expectedData, &expected)
		}

		if unmarshalErr != nil {
			if config.UsePrototext {
				t.Errorf("failed to unmarshal expected prototext from %s: %v", outputFile, unmarshalErr)
			} else {
				t.Errorf("failed to unmarshal expected JSON from %s: %v", outputFile, unmarshalErr)
			}
			return
		}

		if diff := cmp.Diff(expected, result, diffOpts...); diff != "" {
			if *Update {
				var actualData []byte
				var marshalErr error
				
				if config.UsePrototext {
					if resultMsg, ok := any(result).(proto.Message); ok {
						actualData, marshalErr = prototext.MarshalOptions{
							Multiline: true,
							Indent:    "  ",
						}.Marshal(resultMsg)
					} else {
						t.Errorf("result type %T does not implement proto.Message", result)
						return
					}
				} else {
					actualData, marshalErr = json.MarshalIndent(result, "", "  ")
				}
				
				if marshalErr != nil {
					if config.UsePrototext {
						t.Errorf("failed to marshal result to prototext: %v", marshalErr)
					} else {
						t.Errorf("failed to marshal result to JSON: %v", marshalErr)
					}
					return
				}
				if writeErr := os.WriteFile(outputPath, actualData, 0644); writeErr != nil {
					if config.UsePrototext {
						t.Errorf("failed to update prototext output file %s: %v", outputFile, writeErr)
					} else {
						t.Errorf("failed to update JSON output file %s: %v", outputFile, writeErr)
					}
				}
				continue
			}
			t.Errorf("output mismatch for step %d (-expected +got):\n%s", stepNum, diff)
		}
	}
}

// validateAndLoadStepFiles validates that a directory contains a valid sequence of step files
// and loads their content. Returns an error if the sequence is invalid or if any files are unexpected.
func validateAndLoadStepFiles(stepDir, inputExt string, config *TestConfig) ([]StepFile, error) {
	entries, err := os.ReadDir(stepDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read step directory: %w", err)
	}

	var stepFiles []StepFile
	expectedStep := 1

	// Parse and collect all files with the correct extension
	for _, entry := range entries {
		if entry.IsDir() {
			return nil, fmt.Errorf("unexpected subdirectory %s in step directory", entry.Name())
		}

		// Skip output files - check for both success and error output extensions
		if strings.HasSuffix(entry.Name(), config.SuccessOutputExt) || strings.HasSuffix(entry.Name(), config.ErrorOutputExt) {
			continue
		}
		
		if !strings.HasSuffix(entry.Name(), inputExt) {
			return nil, fmt.Errorf("unexpected file %s with wrong extension (expected %s)", entry.Name(), inputExt)
		}

		// Extract step number from filename
		baseName := strings.TrimSuffix(entry.Name(), inputExt)
		stepNum, parseErr := strconv.Atoi(baseName)
		if parseErr != nil {
			return nil, fmt.Errorf("invalid step filename %s: must be a number", entry.Name())
		}

		if stepNum <= 0 {
			return nil, fmt.Errorf("invalid step number %d in filename %s: must be positive", stepNum, entry.Name())
		}

		// Load file content
		filePath := filepath.Join(stepDir, entry.Name())
		data, readErr := os.ReadFile(filePath)
		if readErr != nil {
			return nil, fmt.Errorf("failed to read step file %s: %w", entry.Name(), readErr)
		}

		stepFiles = append(stepFiles, StepFile{
			Step:     stepNum,
			FilePath: filePath,
			Data:     data,
		})
	}

	if len(stepFiles) == 0 {
		return nil, fmt.Errorf("no step files found in directory")
	}

	// Sort by step number
	sort.Slice(stepFiles, func(i, j int) bool {
		return stepFiles[i].Step < stepFiles[j].Step
	})

	// Validate that steps are sequential and dense (no gaps)
	for _, stepFile := range stepFiles {
		if stepFile.Step != expectedStep {
			return nil, fmt.Errorf("step sequence is not dense: expected step %d, found step %d", expectedStep, stepFile.Step)
		}
		expectedStep++
	}

	return stepFiles, nil
}

// RunCombinedTests runs both regular golden file tests and step tests
func RunCombinedTests[T any](t *testing.T, config *TestConfig, testFunc TestFunc[T], stepTestFunc StepTestFunc[T], errorFunc ErrorFunc) {
	// Run regular golden file tests
	RunTests(t, config, testFunc, errorFunc)
	
	// Run step tests
	RunStepTests(t, config, stepTestFunc, errorFunc)
}