package config

import (
	"encoding/json"
	"bytes"
	"os"
	"github.com/google/go-cmp/cmp"
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
				expectedError, readErr := os.ReadFile(filepath.Join(testdataDir, outputFile))
				if readErr != nil {
					t.Fatalf("failed to read expected error output file: %v", readErr)
				}
				if !bytes.Equal(expectedError, []byte(err.Error())) {
					t.Errorf("error output mismatch for file %s:\nExpected:\n%s\nGot:\n%s", file.Name(), expectedError, err.Error())
				}
			} else {
				outputFile += ".out.json"
				if err != nil {
					t.Fatalf("unexpected error for file %s: %v", file.Name(), err)
				}
				expectedConfigData, readErr := os.ReadFile(filepath.Join(testdataDir, outputFile))
				if readErr != nil {
					t.Fatalf("failed to read expected JSON output file: %v", readErr)
				}

				var expectedConfig ServerConfig
				if err := json.Unmarshal(expectedConfigData, &expectedConfig); err != nil {
					t.Fatalf("failed to unmarshal expected JSON: %v", err)
				}

				if diff := cmp.Diff(&expectedConfig, config); diff != "" {
					t.Errorf("config output mismatch for file %s (-expected +got):\n%s", file.Name(), diff)
				}
			}
		})
	}
}
