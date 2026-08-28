package architecture_test

import (
	"bytes"
	"os"
	"strings"
	"testing"
)

const updateArchitectureBaselineEnvironment = "FORTYONE_UPDATE_ARCHITECTURE_BASELINE"

func TestUpdateArchitectureDebtBaseline(t *testing.T) {
	if os.Getenv(updateArchitectureBaselineEnvironment) != "1" {
		t.Skip("architecture baseline updates are explicit; use make architecture-baseline-generate")
	}

	root := serverDir(t)
	findings, err := scanArchitecture(root)
	if err != nil {
		t.Fatalf("scan architecture: %v", err)
	}
	current, _, err := loadArchitectureBaseline(architectureBaselinePath)
	if err != nil {
		t.Fatal(err)
	}
	if growth := compareArchitectureDebt(current, findings); len(growth) > 0 {
		t.Fatalf(
			"refusing to write a baseline that accepts architecture debt growth:\n\n%s",
			strings.Join(growth, "\n\n"),
		)
	}

	baseline := snapshotArchitectureDebt(findings)
	if err := validateArchitectureBaseline(baseline); err != nil {
		t.Fatalf("validate generated architecture baseline: %v", err)
	}
	canonical, err := marshalArchitectureBaseline(baseline)
	if err != nil {
		t.Fatalf("marshal generated architecture baseline: %v", err)
	}
	if err := os.WriteFile(architectureBaselinePath, canonical, 0o644); err != nil {
		t.Fatalf("write architecture baseline: %v", err)
	}

	t.Logf("reviewed architecture debt snapshot written to %s", architectureBaselinePath)
}

func TestArchitectureDebtDoesNotChangeWithoutBaselineReview(t *testing.T) {
	root := serverDir(t)
	findings, err := scanArchitecture(root)
	if err != nil {
		t.Fatalf("scan architecture: %v", err)
	}

	baseline, rawBaseline, err := loadArchitectureBaseline(architectureBaselinePath)
	if err != nil {
		t.Fatal(err)
	}
	canonicalBaseline, err := marshalArchitectureBaseline(baseline)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(rawBaseline, canonicalBaseline) {
		t.Fatalf("%s is not canonical JSON; format it with two-space indentation and one trailing newline", architectureBaselinePath)
	}

	differences := compareArchitectureDebt(baseline, findings)
	if len(differences) > 0 {
		t.Fatalf("architecture debt baseline changed:\n\n%s", strings.Join(differences, "\n\n"))
	}
}
