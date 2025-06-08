package config

import (
	"fmt"
	"net"

	"os"

	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/function"

	hcl "github.com/hashicorp/hcl/v2"
)

type bodyItem string

const (
	blockKind     bodyItem = "block"
	attributeKind bodyItem = "attribute"
)

// Config holds the configuration for the server
type Config struct {
	Server ServerConfig `json:"server"`
}

type ServerConfig struct {
	ListeningAddress string `json:"listening_address"`
}

func ParseConfigFile(filename string) (*Config, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("os.ReadFile(%q): %w", filename, err)
	}

	return ParseConfig(filename, data)
}

func ParseConfig(filename string, src []byte) (*Config, hcl.Diagnostics) {

	beginning := hcl.Pos{Line: 1, Column: 1}

	// ParseServerConfig parses and validates the server configuration from an HCL file
	file, diags := hclsyntax.ParseConfig(src, filename, beginning)
	if diags.HasErrors() {
		return nil, diags
	}

	schema := &hcl.BodySchema{
		Blocks: []hcl.BlockHeaderSchema{
			{
				Type:       "server",
				LabelNames: []string{},
			},
		},
	}

	content, diags := file.Body.Content(schema)
	if diags.HasErrors() {
		return nil, diags
	}

	var config Config
	if len(content.Blocks.OfType("server")) != 1 {
		diags = diags.Append(&hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Expected exactly one server block",
			Detail:   "You provided a non-one number of server blocks.",
		})
	}
	for _, block := range content.Blocks.OfType("server") {
		sc, newDiags := parseServerConfig(block)
		diags = diags.Extend(newDiags)
		config.Server = sc
	}

	return &config, diags
}

func parseServerConfig(block *hcl.Block) (ServerConfig, hcl.Diagnostics) {
	var sc ServerConfig

	schema := &hcl.BodySchema{
		Attributes: []hcl.AttributeSchema{
			{
				Name:     "listening_address",
				Required: true,
			},
		},
	}
	content, diags := block.Body.Content(schema)
	if diags.HasErrors() {
		return ServerConfig{}, diags
	}

	listeningAddress := content.Attributes["listening_address"]
	listeningAddressValue, listeningDiags := listeningAddress.Expr.Value(&hcl.EvalContext{
		Functions: map[string]function.Function{
			"helloworld::with::more::things": function.New(&function.Spec{
				Description: "hello world function",
				Type:        function.StaticReturnType(cty.String),
				Impl: func(args []cty.Value, retType cty.Type) (cty.Value, error) {
					return cty.StringVal("function_with_colons:port"), nil
				},
			}),
		},
	})
	if diags.HasErrors() {
		diags = diags.Extend(listeningDiags)
	} else {
		address := listeningAddressValue.AsString()
		host, port, err := net.SplitHostPort(address)
		if err != nil || host == "" || port == "" {
			diags = diags.Append(&hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  "Invalid listening address",
				Detail:   "The 'listening_address' must be in the format 'host:port'.",
				Subject:  listeningAddress.Expr.Range().Ptr(),
			})
		}
		sc.ListeningAddress = address
	}

	return sc, diags
}
