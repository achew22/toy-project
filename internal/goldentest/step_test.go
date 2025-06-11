package goldentest

import (
	"os"
	"path/filepath"
	"testing"
)

func TestValidateAndLoadStepFiles(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "step_test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	t.Run("valid sequence", func(t *testing.T) {
		// Create valid step files
		stepDir := filepath.Join(tempDir, "valid")
		if err := os.MkdirAll(stepDir, 0755); err != nil {
			t.Fatalf("failed to create step dir: %v", err)
		}

		files := map[string]string{
			"1.hcl": "step 1 content",
			"2.hcl": "step 2 content",
			"3.hcl": "step 3 content",
		}

		for filename, content := range files {
			if err := os.WriteFile(filepath.Join(stepDir, filename), []byte(content), 0644); err != nil {
				t.Fatalf("failed to write file %s: %v", filename, err)
			}
		}

		config := &TestConfig[string]{
			SuccessOutputExt: ".json",
			ErrorOutputExt:   ".txt",
		}
		stepFiles, err := validateAndLoadStepFiles(stepDir, ".hcl", config)
		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}

		if len(stepFiles) != 3 {
			t.Fatalf("expected 3 step files, got %d", len(stepFiles))
		}

		for i, stepFile := range stepFiles {
			expectedStep := i + 1
			if stepFile.Step != expectedStep {
				t.Errorf("expected step %d, got %d", expectedStep, stepFile.Step)
			}
			expectedContent := files[stepFile.FilePath[len(stepFile.FilePath)-5:]] // last 5 chars should be "X.hcl"
			if string(stepFile.Data) != expectedContent {
				t.Errorf("expected content %q, got %q", expectedContent, string(stepFile.Data))
			}
		}
	})

	t.Run("gap in sequence", func(t *testing.T) {
		stepDir := filepath.Join(tempDir, "gap")
		if err := os.MkdirAll(stepDir, 0755); err != nil {
			t.Fatalf("failed to create step dir: %v", err)
		}

		// Create files with a gap (missing step 2)
		files := map[string]string{
			"1.hcl": "step 1 content",
			"3.hcl": "step 3 content",
		}

		for filename, content := range files {
			if err := os.WriteFile(filepath.Join(stepDir, filename), []byte(content), 0644); err != nil {
				t.Fatalf("failed to write file %s: %v", filename, err)
			}
		}

		config := &TestConfig[string]{
			SuccessOutputExt: ".out.json",
			ErrorOutputExt:   ".out.txt",
		}
		_, err := validateAndLoadStepFiles(stepDir, ".hcl", config)
		if err == nil {
			t.Fatal("expected error for gap in sequence, got none")
		}

		if !containsString(err.Error(), "step sequence is not dense") {
			t.Errorf("expected error about dense sequence, got: %v", err)
		}
	})

	t.Run("wrong extension", func(t *testing.T) {
		stepDir := filepath.Join(tempDir, "wrong_ext")
		if err := os.MkdirAll(stepDir, 0755); err != nil {
			t.Fatalf("failed to create step dir: %v", err)
		}

		// Create file with wrong extension
		if err := os.WriteFile(filepath.Join(stepDir, "1.txt"), []byte("content"), 0644); err != nil {
			t.Fatalf("failed to write file: %v", err)
		}

		config := &TestConfig[string]{
			SuccessOutputExt: ".out.json",
			ErrorOutputExt:   ".out.txt",
		}
		_, err := validateAndLoadStepFiles(stepDir, ".hcl", config)
		if err == nil {
			t.Fatal("expected error for wrong extension, got none")
		}

		if !containsString(err.Error(), "unexpected file") {
			t.Errorf("expected error about unexpected file, got: %v", err)
		}
	})

	t.Run("invalid filename", func(t *testing.T) {
		stepDir := filepath.Join(tempDir, "invalid_name")
		if err := os.MkdirAll(stepDir, 0755); err != nil {
			t.Fatalf("failed to create step dir: %v", err)
		}

		// Create file with non-numeric name
		if err := os.WriteFile(filepath.Join(stepDir, "invalid.hcl"), []byte("content"), 0644); err != nil {
			t.Fatalf("failed to write file: %v", err)
		}

		config := &TestConfig[string]{
			SuccessOutputExt: ".out.json",
			ErrorOutputExt:   ".out.txt",
		}
		_, err := validateAndLoadStepFiles(stepDir, ".hcl", config)
		if err == nil {
			t.Fatal("expected error for invalid filename, got none")
		}

		if !containsString(err.Error(), "invalid step filename") {
			t.Errorf("expected error about invalid filename, got: %v", err)
		}
	})

	t.Run("zero step number", func(t *testing.T) {
		stepDir := filepath.Join(tempDir, "zero_step")
		if err := os.MkdirAll(stepDir, 0755); err != nil {
			t.Fatalf("failed to create step dir: %v", err)
		}

		// Create file with step number 0
		if err := os.WriteFile(filepath.Join(stepDir, "0.hcl"), []byte("content"), 0644); err != nil {
			t.Fatalf("failed to write file: %v", err)
		}

		config := &TestConfig[string]{
			SuccessOutputExt: ".out.json",
			ErrorOutputExt:   ".out.txt",
		}
		_, err := validateAndLoadStepFiles(stepDir, ".hcl", config)
		if err == nil {
			t.Fatal("expected error for zero step number, got none")
		}

		if !containsString(err.Error(), "must be positive") {
			t.Errorf("expected error about positive step number, got: %v", err)
		}
	})
}

func containsString(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 || s[len(s)-len(substr):] == substr || s[:len(substr)] == substr || containsString(s[1:], substr))
}

func TestRunStepTests(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "run_step_test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a test step directory
	stepDir := filepath.Join(tempDir, "test_case")
	if err := os.MkdirAll(stepDir, 0755); err != nil {
		t.Fatalf("failed to create step dir: %v", err)
	}

	// Create step files
	files := map[string]string{
		"1.hcl": "first",
		"2.hcl": "second",
	}

	for filename, content := range files {
		if err := os.WriteFile(filepath.Join(stepDir, filename), []byte(content), 0644); err != nil {
			t.Fatalf("failed to write file %s: %v", filename, err)
		}
	}

	// Create expected output files for each step
	step1Output := `first`
	step1OutputFile := filepath.Join(stepDir, "1.out.json")
	if err := os.WriteFile(step1OutputFile, []byte(step1Output), 0644); err != nil {
		t.Fatalf("failed to write step 1 output file: %v", err)
	}

	step2Output := `second`
	step2OutputFile := filepath.Join(stepDir, "2.out.json")
	if err := os.WriteFile(step2OutputFile, []byte(step2Output), 0644); err != nil {
		t.Fatalf("failed to write step 2 output file: %v", err)
	}

	config := &TestConfig[string]{
		InputExt:         ".hcl",
		ErrorOutputExt:   ".txt",
		SuccessOutputExt: ".json",
		StepTestFunc: func(stepFile StepFile) (string, error) {
			return string(stepFile.Data), nil
		},
		ErrorFunc: func(err error) []byte {
			return []byte(err.Error())
		},
	}

	// This should pass without errors
	config.RunTests(t, tempDir)
}

// Helper function that mimics strings.Contains for basic substring checking
func strings_Join(elems []string, sep string) string {
	switch len(elems) {
	case 0:
		return ""
	case 1:
		return elems[0]
	}

	var result string
	for i, elem := range elems {
		if i > 0 {
			result += sep
		}
		result += elem
	}
	return result
}
