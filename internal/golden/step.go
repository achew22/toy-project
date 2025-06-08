package golden

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"testing"
)

// StepTestFunc is a function that processes a sequence of input files and returns either a result or an error
type StepTestFunc[T any] func(stepFiles []StepFile) (T, error)

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
			stepFiles, validateErr := validateAndLoadStepFiles(stepDir, config.InputExt)
			if validateErr != nil {
				t.Fatalf("failed to validate step directory %s: %v", entry.Name(), validateErr)
			}

			result, testErr := stepTestFunc(stepFiles)

			if strings.HasPrefix(entry.Name(), config.ErrorPrefix) {
				// This is an error test case
				testErrorCase[T](t, config, entry.Name(), entry.Name(), testErr, errorFunc)
			} else {
				// This is a success test case
				testSuccessCase[T](t, config, entry.Name(), entry.Name(), result, testErr)
			}
		})
	}
}

// validateAndLoadStepFiles validates that a directory contains a valid sequence of step files
// and loads their content. Returns an error if the sequence is invalid or if any files are unexpected.
func validateAndLoadStepFiles(stepDir, inputExt string) ([]StepFile, error) {
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

		if filepath.Ext(entry.Name()) != inputExt {
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