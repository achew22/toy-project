package config

import (
	"fmt"
	"net"

	"github.com/hashicorp/hcl/v2/gohcl"
	"github.com/hashicorp/hcl/v2/hclparse"
)

// ServerConfig holds the configuration for the server
type ServerConfig struct {
	ListeningAddress string `hcl:"listening_address"`
}

// ParseServerConfig parses and validates the server configuration from an HCL file
func ParseServerConfig(filename string) (*ServerConfig, error) {
	parser := hclparse.NewParser()
	file, diags := parser.ParseHCLFile(filename)
	if diags.HasErrors() {
		return nil, fmt.Errorf("failed to parse HCL file: %s", diags.Error())
	}

	var config ServerConfig
	diags = gohcl.DecodeBody(file.Body, nil, &config)
	if diags.HasErrors() {
		return nil, fmt.Errorf("failed to decode HCL body: %s", diags.Error())
	}

	if err := validateServerConfig(&config); err != nil {
		return nil, err
	}

	return &config, nil
}

// validateServerConfig validates the server configuration
func validateServerConfig(config *ServerConfig) error {
	if config.ListeningAddress == "" {
		return fmt.Errorf("listening_address must be specified")
	}

	host, port, err := net.SplitHostPort(config.ListeningAddress)
	if err != nil {
		return fmt.Errorf("invalid listening_address format: %v", err)
	}

	if net.ParseIP(host) == nil {
		return fmt.Errorf("invalid IP address: %s", host)
	}

	if _, err := net.LookupPort("tcp", port); err != nil {
		return fmt.Errorf("invalid port: %s", port)
	}

	return nil
}
