package main

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
	"testing/fstest"
)

func TestAuditAllowsOnlyStandaloneCanonicalWebResponses(t *testing.T) {
	t.Parallel()

	const source = `package fixture
import transport "github.com/complexus-tech/projects-api/pkg/web"
func handler() {
	transport.Respond(nil, nil, nil, 200)
	transport.RespondError(nil, nil, nil, 500)
}
`
	report := encodeReport(t, []gosecIssue{
		{RuleID: "G104", File: "/repo/handler.go", Line: "4", Column: "2"},
		{RuleID: "G104", File: "/repo/handler.go", Line: "5", Column: "2"},
	})

	result, err := auditGosec(strings.NewReader(report), "/repo", fstest.MapFS{
		"handler.go": &fstest.MapFile{Data: []byte(source)},
	})
	if err != nil {
		t.Fatalf("audit G104: %v", err)
	}
	if result.Allowed != 2 || len(result.Rejected) != 0 {
		t.Fatalf("result = %+v, want two allowed response writes", result)
	}
}

func TestAuditRejectsEveryOtherUncheckedErrorShape(t *testing.T) {
	t.Parallel()

	const source = `package fixture
import transport "github.com/complexus-tech/projects-api/pkg/web"
type localWeb struct{}
func (localWeb) Respond(any, any, any, int) error { return nil }
func cleanup() error { return nil }
func handler() {
	cleanup()
	_ = transport.Respond(nil, nil, nil, 200)
	localWeb{}.Respond(nil, nil, nil, 200)
}
`
	report := encodeReport(t, []gosecIssue{
		{RuleID: "G104", File: "/repo/handler.go", Line: "7", Column: "2"},
		{RuleID: "G104", File: "/repo/handler.go", Line: "8", Column: "2"},
		{RuleID: "G104", File: "/repo/handler.go", Line: "9", Column: "2"},
	})

	result, err := auditGosec(strings.NewReader(report), "/repo", fstest.MapFS{
		"handler.go": &fstest.MapFile{Data: []byte(source)},
	})
	if err != nil {
		t.Fatalf("audit G104: %v", err)
	}
	if result.Allowed != 0 || len(result.Rejected) != 3 {
		t.Fatalf("result = %+v, want three rejected findings", result)
	}
}

func TestAuditRejectsShadowedCanonicalImport(t *testing.T) {
	t.Parallel()

	const source = `package fixture
import transport "github.com/complexus-tech/projects-api/pkg/web"
type localWeb struct{}
func (localWeb) Respond(any, any, any, int) error { return nil }
func handler() {
	transport := localWeb{}
	transport.Respond(nil, nil, nil, 200)
}
`
	report := encodeReport(t, []gosecIssue{
		{RuleID: "G104", File: "/repo/handler.go", Line: "7", Column: "2"},
	})

	result, err := auditGosec(strings.NewReader(report), "/repo", fstest.MapFS{
		"handler.go": &fstest.MapFile{Data: []byte(source)},
	})
	if err != nil {
		t.Fatalf("audit G104: %v", err)
	}
	if result.Allowed != 0 || len(result.Rejected) != 1 {
		t.Fatalf("result = %+v, want the shadowed import rejected", result)
	}
}

func TestAuditRejectsEveryNonG104FindingWithoutReadingSource(t *testing.T) {
	t.Parallel()

	report := encodeReport(t, []gosecIssue{
		{RuleID: "G101", File: "/repo/missing.go", Line: "8", Column: "4"},
	})
	result, err := auditGosec(strings.NewReader(report), "/repo", fstest.MapFS{})
	if err != nil {
		t.Fatalf("audit gosec: %v", err)
	}
	if result.Allowed != 0 || len(result.Rejected) != 1 || result.Rejected[0].RuleID != "G101" {
		t.Fatalf("result = %+v, want the non-G104 finding rejected", result)
	}
}

func TestAuditAllowsACompleteReportWithNoFindings(t *testing.T) {
	t.Parallel()

	result, err := auditGosec(strings.NewReader(encodeReport(t, nil)), "/repo", fstest.MapFS{})
	if err != nil {
		t.Fatalf("audit empty gosec report: %v", err)
	}
	if result.Allowed != 0 || len(result.Rejected) != 0 {
		t.Fatalf("result = %+v, want an empty successful audit", result)
	}
}

func TestAuditRejectsDuplicateScannerLocations(t *testing.T) {
	t.Parallel()

	report := encodeReport(t, []gosecIssue{
		{RuleID: "G101", File: "/repo/handler.go", Line: "3", Column: "2"},
		{RuleID: "G101", File: "/repo/handler.go", Line: "3", Column: "2"},
	})
	_, err := auditGosec(strings.NewReader(report), "/repo", fstest.MapFS{})
	if err == nil || !strings.Contains(err.Error(), "duplicate finding") {
		t.Fatalf("error = %v, want duplicate scanner location rejected", err)
	}
}

func TestAuditRejectsPathsOutsideSourceRoot(t *testing.T) {
	t.Parallel()

	report := encodeReport(t, []gosecIssue{
		{RuleID: "G101", File: "/elsewhere/handler.go", Line: "3", Column: "2"},
	})
	_, err := auditGosec(strings.NewReader(report), "/repo", fstest.MapFS{})
	if err == nil || !strings.Contains(err.Error(), "outside the source root") {
		t.Fatalf("error = %v, want path escape rejected", err)
	}
}

func TestAuditRejectsTrailingScannerData(t *testing.T) {
	t.Parallel()

	report := encodeReport(t, nil) + ` {"unexpected":true}`
	_, err := auditGosec(strings.NewReader(report), "/repo", fstest.MapFS{})
	if err == nil || !strings.Contains(err.Error(), "trailing data") {
		t.Fatalf("error = %v, want trailing scanner data rejected", err)
	}
}

func TestAuditRejectsIncompleteScannerOutput(t *testing.T) {
	t.Parallel()

	report := `{"Golang errors":{"fixture":{"error":"sensitive compiler output"}},"Issues":[]}`
	_, err := auditGosec(strings.NewReader(report), "/repo", fstest.MapFS{})
	if err == nil || strings.Contains(err.Error(), "sensitive") {
		t.Fatalf("error = %v, want redacted scanner failure", err)
	}
}

func TestAuditRejectsScannerOutputThatAnalyzedNoFiles(t *testing.T) {
	t.Parallel()

	report := `{"Golang errors":{},"Issues":[],"Stats":{"files":0}}`
	_, err := auditGosec(strings.NewReader(report), "/repo", fstest.MapFS{})
	if err == nil || !strings.Contains(err.Error(), "did not analyze any files") {
		t.Fatalf("error = %v, want empty scanner report rejected", err)
	}
}

func TestAuditRejectsInvalidRuleIdentifierWithoutEchoingIt(t *testing.T) {
	t.Parallel()

	report := `{"Golang errors":{},"Issues":[{"rule_id":"secret-value","file":"/repo/handler.go","line":"3","column":"18"}],"Stats":{"files":1}}`
	_, err := auditGosec(strings.NewReader(report), "/repo", fstest.MapFS{})
	if err == nil || strings.Contains(err.Error(), "secret-value") {
		t.Fatalf("error = %v, want redacted invalid rule failure", err)
	}
}

func TestRejectedOutputNeverIncludesGosecSourceSnippet(t *testing.T) {
	t.Parallel()

	const source = "package fixture\nfunc cleanup() error { return nil }\nfunc handler() { cleanup() }\n"
	report := `{"Golang errors":{},"Issues":[{"rule_id":"G104","file":"/repo/handler.go","line":"3","column":"18","code":"super-secret-bearer-token"}],"Stats":{"files":1}}`
	result, err := auditGosec(strings.NewReader(report), "/repo", fstest.MapFS{
		"handler.go": &fstest.MapFile{Data: []byte(source)},
	})
	if err != nil {
		t.Fatalf("audit G104: %v", err)
	}
	var output bytes.Buffer
	if err := writeAudit(&output, result); err != nil {
		t.Fatalf("write audit: %v", err)
	}
	if strings.Contains(output.String(), "super-secret") {
		t.Fatalf("output leaked gosec source: %s", output.String())
	}
	if !strings.Contains(output.String(), "handler.go:3:18") {
		t.Fatalf("output omitted safe location: %s", output.String())
	}
}

func encodeReport(t *testing.T, issues []gosecIssue) string {
	t.Helper()
	contents, err := json.Marshal(gosecReport{
		GolangErrors: map[string]json.RawMessage{},
		Issues:       issues,
		Stats:        gosecStats{Files: 1},
	})
	if err != nil {
		t.Fatalf("encode report: %v", err)
	}
	return string(contents)
}
