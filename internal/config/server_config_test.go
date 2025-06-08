package config

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

var update = flag.Bool("update", false, "update .out files if there is a difference")

func TestParseConfig(t *testing.T) {
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

			outputFile := strings.TrimSuffix(file.Name(), ".hcl")

			config, diags := ParseConfig(filePath, data)
			if strings.HasPrefix(file.Name(), "error_") {
				outputFile += ".out.txt"
				if diags == nil {
					t.Errorf("expected error for file %s, but got none", file.Name())
				}
				expectedError, readErr := os.ReadFile(filepath.Join(testdataDir, outputFile))
				if readErr != nil {
					t.Logf("failed to read expected error output file: %v", readErr)
				}

				var actualError []byte
				if diags.HasErrors() {
					actualError = []byte(diags.Error())
				}
				if !bytes.Equal(expectedError, actualError) {
					if *update {
						if writeErr := os.WriteFile(filepath.Join(testdataDir, outputFile), actualError, 0644); writeErr != nil {
							t.Errorf("failed to update error output file: %v", writeErr)
						}
					}
					t.Errorf("error output mismatch for file %s:\nExpected:\n%s\nGot:\n%s", file.Name(), expectedError, actualError)
				}
			} else {
				outputFile += ".out.json"
				if diags.HasErrors() {
					t.Errorf("unexpected error for file %s: %v", file.Name(), diags)
				}
				expectedConfigData, readErr := os.ReadFile(filepath.Join(testdataDir, outputFile))
				if readErr != nil {
					t.Logf("failed to read expected JSON output file: %v", readErr)
				}

				var expectedConfig Config
				if err := json.Unmarshal(expectedConfigData, &expectedConfig); err != nil {
					t.Errorf("failed to unmarshal expected JSON: %v", err)
				}

				if diff := cmp.Diff(&expectedConfig, config, cmpopts.EquateEmpty()); diff != "" {
					if *update {
						actualConfigData, marshalErr := json.MarshalIndent(config, "", "  ")
						if marshalErr != nil {
							t.Errorf("failed to marshal config to JSON: %v", marshalErr)
						}
						if writeErr := os.WriteFile(filepath.Join(testdataDir, outputFile), actualConfigData, 0644); writeErr != nil {
							t.Errorf("failed to update JSON output file: %v", writeErr)
						}
					}
					t.Errorf("config output mismatch for file %s (-expected +got):\n%s", file.Name(), diff)
				}
			}
		})
	}
}
