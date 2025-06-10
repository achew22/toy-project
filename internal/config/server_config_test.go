package config

import (
	"testing"

	"github.com/achew22/toy-project/internal/goldentest"
	"github.com/hashicorp/hcl/v2"
)

func TestParseConfig(t *testing.T) {
	config := goldentest.DefaultConfig()
	
	testFunc := func(filePath string, data []byte) (*Config, error) {
		config, diags := ParseConfig(filePath, data)
		if diags.HasErrors() {
			return nil, diags
		}
		return config, nil
	}
	
	errorFunc := func(err error) []byte {
		if diags, ok := err.(hcl.Diagnostics); ok && diags.HasErrors() {
			return []byte(diags.Error())
		}
		return []byte(err.Error())
	}

	goldentest.RunTests(t, config, testFunc, errorFunc)
}
