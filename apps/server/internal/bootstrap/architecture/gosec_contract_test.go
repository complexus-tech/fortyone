package architecture_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGosecG104ExceptionRemainsNarrowAndGoverned(t *testing.T) {
	t.Parallel()

	root := serverDir(t)
	makefile, err := os.ReadFile(filepath.Join(root, "Makefile"))
	if err != nil {
		t.Fatalf("read server Makefile: %v", err)
	}
	contract := string(makefile)
	for _, required := range []string{
		"GOSEC_VERSION := v2.29.0",
		"@$(MAKE) g104-check",
		"g104-check:",
		"@go run github.com/securego/gosec/v2/cmd/gosec@$(GOSEC_VERSION) -exclude-generated -fmt=json ./... 2>/dev/null | go run ./internal/tools/g104guard -root .",
	} {
		if !strings.Contains(contract, required) {
			t.Errorf("gosec contract is missing %q", required)
		}
	}
	for _, forbidden := range []string{"-exclude=G104", "-include=G104", "-quiet -exclude-generated -fmt=json"} {
		if strings.Contains(contract, forbidden) {
			t.Errorf("gosec contract contains unsafe scanner option %q", forbidden)
		}
	}

	workflow, err := os.ReadFile(filepath.Join(root, "..", "..", ".github", "workflows", "weekly-assurance.yml"))
	if err != nil {
		t.Fatalf("read weekly assurance workflow: %v", err)
	}
	if !strings.Contains(string(workflow), "run: make security-check") {
		t.Fatal("weekly assurance workflow does not execute the governed security gate")
	}

	documentation, err := os.ReadFile(filepath.Join(root, "docs", "security", "quality-gates.md"))
	if err != nil {
		t.Fatalf("read quality gate documentation: %v", err)
	}
	for _, required := range []string{
		"## Governed G104 transport exception",
		"internal/tools/g104guard",
		"not blanket-suppressed",
	} {
		if !strings.Contains(string(documentation), required) {
			t.Errorf("quality gate documentation is missing %q", required)
		}
	}
}
