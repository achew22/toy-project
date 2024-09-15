package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestParseServerConfig(t *testing.T) {
	testdataDir := "testdata"
	files, err := os.ReadDir(testdataDir)
	if err != nil {
		t.Fatalf("failed to read testdata directory: %v", err)
	}

	for _, file := range files {
		if filepath.Ext(file.Name()) != ".hcl" {
			continue
		}

		t.Run(file.Name(), func(t *testing.T) {
			filePath := filepath.Join(testdataDir, file.Name())
			data, err := os.ReadFile(filePath)
			if err != nil {
				t.Fatalf("failed to read file %s: %v", file.Name(), err)
			}

			config, err := ParseServerConfig(filePath, data)
			outputFile := strings.TrimSuffix(file.Name(), ".hcl")

			if strings.HasPrefix(file.Name(), "error_") {
				outputFile += ".out.txt"
				if err == nil {
					t.Fatalf("expected error for file %s, but got none", file.Name())
				}
				err = os.WriteFile(filepath.Join(testdataDir, outputFile), []byte(err.Error()), 0644)
				if err != nil {
					t.Fatalf("failed to write error output file: %v", err)
				}
			} else {
				outputFile += ".out.json"
				if err != nil {
					t.Fatalf("unexpected error for file %s: %v", file.Name(), err)
				}
				jsonData, err := json.MarshalIndent(config, "", "  ")
				if err != nil {
					t.Fatalf("failed to marshal config to JSON: %v", err)
				}
				err = os.WriteFile(filepath.Join(testdataDir, outputFile), jsonData, 0644)
				if err != nil {
					t.Fatalf("failed to write JSON output file: %v", err)
				}
			}
		})
	}
}
