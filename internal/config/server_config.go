package config

import (
	"fmt"
	"net"

	"bufio"
	"os"
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclparse"
)

// ServerConfig holds the configuration for the server
type ServerConfig struct {
	ListeningAddress string `hcl:"listening_address"`
}

// ParseServerConfig parses and validates the server configuration from an HCL file
func ParseServerConfig(filename string) (*ServerConfig, error) {
	parser := hclparse.NewParser()
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to open HCL file: %v", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var config ServerConfig
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "listening_address") {
			parts := strings.Split(line, "=")
			if len(parts) != 2 {
				return nil, fmt.Errorf("invalid format for listening_address")
			}
			config.ListeningAddress = strings.TrimSpace(strings.Trim(parts[1], `"`))
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading HCL file: %v", err)
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
