package goldentest

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

// StepTestFunc is a function that processes a single input file and returns either a result or an error.
// The function receives a context, fixture, and StepFile containing the step information.
//
// Parameters:
//   - ctx: Context for the test step (for cancellation and timeouts)
//   - fixture: Fixture created by SetUpFunc for this test case
//   - stepFile: Information about the current step including file path and content
//
// Returns:
//   - T: Result of processing the step (for success cases)
//   - error: Error encountered during processing (for error cases)
//
// Example:
//
//	StepTestFunc: func(ctx context.Context, fixture *ServerFixture, stepFile StepFile) (*Response, error) {
//		return fixture.Client.ProcessStep(ctx, stepFile.Data)
//	}
type StepTestFunc[T, F any] func(ctx context.Context, fixture F, stepFile StepFile) (T, error)

// StepFile represents a single step in a sequence with its file path and data
type StepFile struct {
	// Step is the 1-based step number
	Step int
	// FilePath is the full path to the step file
	FilePath string
	// Data is the content of the step file
	Data []byte
}

// runStepTests runs golden file tests in step mode for all directories in the specified directory
func (config *TestConfig[T, F]) runStepTests(t *testing.T, dir string) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("failed to read testdata directory: %v", err)
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		t.Run(entry.Name(), func(t *testing.T) {
			// Set up fixture for this test case
			var fixture F
			var setUpErr error
			if config.SetUp != nil {
				fixture, setUpErr = config.SetUp(t)
				if setUpErr != nil {
					t.Fatalf("SetUp failed for test %s: %v", entry.Name(), setUpErr)
				}
			}

			// Ensure teardown runs even if test fails
			defer func() {
				if config.TearDown != nil {
					if tearDownErr := config.TearDown(t, fixture); tearDownErr != nil {
						t.Errorf("TearDown failed for test %s: %v", entry.Name(), tearDownErr)
					}
				}
			}()

			stepDir := filepath.Join(dir, entry.Name())
			stepFiles, validateErr := validateAndLoadStepFiles[T, F](stepDir, config)
			if validateErr != nil {
				t.Fatalf("failed to validate step directory %s: %v", entry.Name(), validateErr)
			}

			var results []T
			var testErr error

			// Execute stepTestFunc for each step file
			for _, stepFile := range stepFiles {
				result, err := config.StepTestFunc(t.Context(), fixture, stepFile)
				if err != nil {
					testErr = err
					break
				}
				results = append(results, result)
			}

			// Check if error handling is configured
			errorHandlingEnabled := config.ErrorFunc != nil

			if errorHandlingEnabled && strings.HasPrefix(entry.Name(), config.ErrorPrefix) {
				// This is an error test case - only test the final result
				if testErr == nil {
					t.Errorf("expected error for test %s, but got none", entry.Name())
					return
				}
				// Test error with final step number as filename
				finalStepNum := len(stepFiles)
				errorFile := fmt.Sprintf("%d.out%s", finalStepNum, config.ErrorOutputExt)
				config.testErrorCaseStep(t, stepDir, errorFile, testErr, config.ErrorFunc)
			} else {
				// This is a success test case (or error handling is disabled)
				if testErr != nil {
					if !errorHandlingEnabled {
						t.Errorf("test failed for %s: %v", entry.Name(), testErr)
						return
					}
					// Error handling is enabled but this isn't an error test case
					t.Errorf("unexpected error for test %s: %v", entry.Name(), testErr)
					return
				}
				config.testSuccessCaseSteps(t, stepDir, stepFiles, results)
			}
		})
	}
}

func (config *TestConfig[T, F]) testErrorCaseStep(t *testing.T, stepDir, errorFile string, testErr error, errorFunc ErrorFunc) {
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

func (config *TestConfig[T, F]) testSuccessCaseSteps(t *testing.T, stepDir string, stepFiles []StepFile, results []T) {
	if len(results) != len(stepFiles) {
		t.Errorf("expected %d results, got %d", len(stepFiles), len(results))
		return
	}

	var diffOpts []cmp.Option
	diffOpts = append(diffOpts, cmpopts.EquateEmpty())
	diffOpts = append(diffOpts, config.DiffOpts...)

	for i, result := range results {
		stepNum := stepFiles[i].Step
		outputFile := fmt.Sprintf("%d.out%s", stepNum, config.SuccessOutputExt)
		outputPath := filepath.Join(stepDir, outputFile)

		expectedData, readErr := os.ReadFile(outputPath)
		if readErr != nil {
			t.Logf("failed to read expected output file %s: %v", outputFile, readErr)
		}

		// Load expected value from golden file
		expected, loadErr := config.Loader(expectedData)
		if loadErr != nil {
			t.Errorf("failed to load expected value from %s: %v", outputFile, loadErr)
			return
		}

		if diff := cmp.Diff(expected, result, diffOpts...); diff != "" {
			if *Update {
				// Format the actual result for writing to golden file
				actualData, formatErr := config.Formatter(result)
				if formatErr != nil {
					t.Errorf("failed to format result for step %d: %v", stepNum, formatErr)
					return
				}

				if writeErr := os.WriteFile(outputPath, actualData, 0644); writeErr != nil {
					t.Errorf("failed to update output file %s: %v", outputFile, writeErr)
				}
				continue
			}
			t.Errorf("output mismatch for step %d (-expected +got):\n%s", stepNum, diff)
		}
	}
}

// validateAndLoadStepFiles validates that a directory contains a valid sequence of step files
// and loads their content. Returns an error if the sequence is invalid or if any files are unexpected.
func validateAndLoadStepFiles[T, F any](stepDir string, config *TestConfig[T, F]) ([]StepFile, error) {
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

		// Skip output files - check for both success and error output extensions with .out prefix
		if strings.HasSuffix(entry.Name(), ".out"+config.SuccessOutputExt) || strings.HasSuffix(entry.Name(), ".out"+config.ErrorOutputExt) {
			continue
		}

		if !strings.HasSuffix(entry.Name(), ".in"+config.InputExt) {
			return nil, fmt.Errorf("unexpected file %s with wrong extension (expected %s)", entry.Name(), ".in"+config.InputExt)
		}

		// Extract step number from filename
		baseName := strings.TrimSuffix(entry.Name(), ".in"+config.InputExt)
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
