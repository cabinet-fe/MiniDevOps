package main

import (
	"bytes"
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

// openapi-project regenerates api/openapi.3.1.projection.yaml from api/openapi.yaml.
// It downgrades the openapi version field to 3.1.0 for tooling that lacks 3.2 support.
// Never hand-edit the projection; always regenerate via `make openapi-projection`.
func main() {
	srcPath := "api/openapi.yaml"
	dstPath := "api/openapi.3.1.projection.yaml"
	if len(os.Args) >= 3 {
		srcPath = os.Args[1]
		dstPath = os.Args[2]
	}

	raw, err := os.ReadFile(srcPath)
	if err != nil {
		fatalf("read source: %v", err)
	}

	var doc yaml.Node
	if err := yaml.Unmarshal(raw, &doc); err != nil {
		fatalf("parse source: %v", err)
	}

	if err := setOpenAPIVersion(&doc, "3.1.0"); err != nil {
		fatalf("%v", err)
	}

	var buf bytes.Buffer
	enc := yaml.NewEncoder(&buf)
	enc.SetIndent(2)
	if err := enc.Encode(&doc); err != nil {
		fatalf("encode projection: %v", err)
	}
	_ = enc.Close()

	header := "# GENERATED FILE — DO NOT EDIT.\n# Source: api/openapi.yaml (OpenAPI 3.2)\n# Regenerate: make openapi-projection\n"
	out := header + buf.String()

	if err := os.WriteFile(dstPath, []byte(out), 0o644); err != nil {
		fatalf("write projection: %v", err)
	}
	fmt.Printf("wrote %s\n", dstPath)
}

func setOpenAPIVersion(doc *yaml.Node, version string) error {
	root := doc
	if root.Kind == yaml.DocumentNode && len(root.Content) > 0 {
		root = root.Content[0]
	}
	if root.Kind != yaml.MappingNode {
		return fmt.Errorf("openapi root is not a mapping")
	}
	for i := 0; i+1 < len(root.Content); i += 2 {
		k := root.Content[i]
		if k.Value == "openapi" {
			root.Content[i+1].Value = version
			root.Content[i+1].Tag = "!!str"
			return nil
		}
	}
	return fmt.Errorf("openapi version field not found")
}

func fatalf(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	if !strings.HasSuffix(msg, "\n") {
		msg += "\n"
	}
	fmt.Fprint(os.Stderr, msg)
	os.Exit(1)
}
