package config

import (
	"fmt"

	"os"

	hcl "github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
)

// ServerConfig holds the configuration for the server
type ServerConfig struct {
	ListeningAddress string `json:"listening_address"`
}

func ParseServerConfigFile(filename string) (*ServerConfig, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("os.ReadFile(%q): %w", filename, err)
	}

	return ParseServerConfig(filename, data)
}

func ParseServerConfig(filename string, src []byte) (*ServerConfig, error) {

	// ParseServerConfig parses and validates the server configuration from an HCL file
	file, diags := hclsyntax.ParseConfig(src, filename, hcl.Pos{Line: 1, Column: 1})
	if diags.HasErrors() {
		return nil, fmt.Errorf("failed to parse HCL file: %s", diags.Error())
	}

	body, ok := file.Body.(*hclsyntax.Body)
	if !ok {
		return nil, fmt.Errorf("unexpected HCL body type")
	}

	var config ServerConfig
	for _, block := range body.Blocks {
		if block.Type == "server" {
			for _, attr := range block.Body.Attributes {
				if attr.Name == "listening_address" {
					value, diags := attr.Expr.Value(nil)
					if diags.HasErrors() {
						return nil, fmt.Errorf("error reading listening_address: %s", diags.Error())
					}
					config.ListeningAddress = value.AsString()
				}
			}
		}
	}

	return &config, nil
}
