// Command openapibundle validates and internalizes the checked-in split
// OpenAPI source before code generation. The split files remain the source of
// truth; the temporary bundle is deliberately never committed.
package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
)

func main() {
	input := flag.String("input", "api/openapi/v1/openapi.yaml", "root OpenAPI document")
	output := flag.String("output", "", "destination for the internalized JSON bundle")
	flag.Parse()
	if flag.NArg() != 0 || *output == "" {
		fatal(errors.New("usage: openapibundle -input <root.yaml> -output <bundle.json>"))
	}

	loader := openapi3.NewLoader()
	loader.IsExternalRefsAllowed = true
	document, err := loader.LoadFromFile(*input)
	if err != nil {
		fatal(fmt.Errorf("load split OpenAPI document: %w", err))
	}
	ctx := context.Background()
	document.InternalizeRefs(ctx, stableComponentName)
	if err := document.Validate(ctx); err != nil {
		fatal(fmt.Errorf("validate OpenAPI document: %w", err))
	}

	bundle, err := json.MarshalIndent(document, "", "  ")
	if err != nil {
		fatal(fmt.Errorf("encode OpenAPI bundle: %w", err))
	}
	bundle = append(bundle, '\n')
	if err := os.MkdirAll(filepath.Dir(*output), 0o750); err != nil {
		fatal(fmt.Errorf("create OpenAPI bundle directory: %w", err))
	}
	if err := os.WriteFile(*output, bundle, 0o600); err != nil {
		fatal(fmt.Errorf("write OpenAPI bundle: %w", err))
	}
}

// stableComponentName keeps generated Go names independent of the directory
// layout used to split the source contract. Component names are unique within
// this API by policy and are reviewed by the generator drift gate.
func stableComponentName(document *openapi3.T, ref openapi3.ComponentRef) string {
	if parsed := ref.RefPath(); parsed != nil && path.Base(parsed.Path) == "openapi.yaml" {
		fragment := strings.Trim(parsed.Fragment, "/")
		parts := strings.Split(fragment, "/")
		if len(parts) >= 3 && parts[0] == "components" {
			if name := path.Base(fragment); name != "." && name != "/" && name != "" {
				return name
			}
		}
	}
	return openapi3.DefaultRefNameResolver(document, ref)
}

func fatal(err error) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}
