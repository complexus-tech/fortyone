package architecture_test

import (
	"go/token"
	"strings"
	"testing"
)

func TestInspectGoSourceFindsConcreteLayerDependencies(t *testing.T) {
	t.Parallel()

	repositorySource := []byte(`package storiesrepository

import (
	commentsrepository "github.com/complexus-tech/projects-api/internal/modules/comments/repository"
	storiesql "github.com/complexus-tech/projects-api/internal/modules/stories/repository/sqlc"
	stories "github.com/complexus-tech/projects-api/internal/modules/stories/service"
)

var _ = commentsrepository.Repository{}
var _ = storiesql.Queries{}
var _ stories.CoreStory
`)
	findings := inspectFixture(t, "internal/modules/stories/repository/models.go", repositorySource)
	assertRuleCount(t, findings, ruleRepositoryImport, 1)
	assertRuleCount(t, findings, ruleRepositoryServiceImport, 1)
	assertRuleCount(t, findings, ruleGeneratedSQLCLeak, 0)

	serviceSource := []byte(`package storiesservice

import (
	comments "github.com/complexus-tech/projects-api/internal/modules/comments/service"
	storiesrepository "github.com/complexus-tech/projects-api/internal/modules/stories/repository"
	openapi "github.com/complexus-tech/projects-api/api/openapi/v1/generated"
)

var _ = comments.Service{}
var _ = storiesrepository.Repository{}
var _ openapi.Story
`)
	findings = inspectFixture(t, "internal/modules/stories/service/stories.go", serviceSource)
	assertRuleCount(t, findings, ruleServiceRepositoryImport, 1)
	assertRuleCount(t, findings, ruleCrossModuleServiceImport, 1)
	assertRuleCount(t, findings, ruleGeneratedOpenAPILeak, 1)
}

func TestInspectGoSourceAllowsPortsAndKeepsGeneratedOpenAPIAtTransport(t *testing.T) {
	t.Parallel()

	source := []byte(`package storieshttp

import (
	"github.com/complexus-tech/projects-api/api/openapi/v1/generated"
	"github.com/complexus-tech/projects-api/internal/modules/stories/service"
)

var _ generated.Story
var _ service.StoryReader
`)
	findings := inspectFixture(t, "internal/modules/stories/http/routes.go", source)
	assertRuleCount(t, findings, ruleGeneratedOpenAPILeak, 0)
	assertRuleCount(t, findings, ruleCrossModuleServiceImport, 0)
}

func TestInspectGoSourceKeepsGeneratedOpenAPIOutOfModuleDomain(t *testing.T) {
	t.Parallel()

	source := []byte(`package stories

import openapi "github.com/complexus-tech/projects-api/api/openapi/v1/generated"

var _ openapi.Story
`)
	findings := inspectFixture(t, "internal/modules/stories/model.go", source)
	assertRuleCount(t, findings, ruleGeneratedOpenAPILeak, 1)
}

func TestInspectGoSourceFindsRepositoryAuthContextReads(t *testing.T) {
	t.Parallel()

	source := []byte(`package reportsrepository

import (
	"context"
	identity "github.com/complexus-tech/projects-api/internal/platform/auth"
)

func Load(ctx context.Context) error {
	_, err := identity.GetUserID(ctx)
	return err
}
`)
	findings := inspectFixture(t, "internal/modules/reports/repository/queries.go", source)
	assertRuleCount(t, findings, ruleRepositoryAuthContextRead, 1)
	if detail := findingsForRule(findings, ruleRepositoryAuthContextRead)[0].Detail; !strings.Contains(detail, "GetUserID") {
		t.Fatalf("auth-context finding is not actionable: %s", detail)
	}

	serviceSource := []byte(`package reportsservice

import (
	"context"
	identity "github.com/complexus-tech/projects-api/internal/platform/auth"
)

func Load(ctx context.Context) error {
	_, err := identity.GetUserID(ctx)
	return err
}
`)
	findings = inspectFixture(t, "internal/modules/reports/service/reports.go", serviceSource)
	assertRuleCount(t, findings, ruleRepositoryAuthContextRead, 0)
}

func TestInspectGoSourceFindsDotImportedRepositoryAuthContextReads(t *testing.T) {
	t.Parallel()

	source := []byte(`package reportsrepository

import (
	"context"
	. "github.com/complexus-tech/projects-api/internal/platform/auth"
)

func Load(ctx context.Context) error {
	_, err := GetActor(ctx)
	return err
}
`)
	findings := inspectFixture(t, "internal/modules/reports/repository/queries.go", source)
	assertRuleCount(t, findings, ruleRepositoryAuthContextRead, 1)
}

func TestInspectGoSourceRejectsUnsafeRequestBodyReads(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		body    string
		imports string
		want    int
	}{
		{
			name:    "unbounded read all",
			imports: `"io"`,
			body:    `_, _ = io.ReadAll(r.Body)`,
			want:    1,
		},
		{
			name:    "unbounded aliased body",
			imports: `"io"`,
			body: `body := r.Body
	_, _ = io.ReadAll(body)`,
			want: 1,
		},
		{
			name:    "truncating limit reader",
			imports: `"io"`,
			body:    `_, _ = io.ReadAll(io.LimitReader(r.Body, 1024))`,
			want:    1,
		},
		{
			name:    "bounded helper",
			imports: `web "github.com/complexus-tech/projects-api/pkg/web"`,
			body:    `_, _ = web.ReadBoundedBody(w, r, 1024)`,
			want:    0,
		},
		{
			name:    "rejecting max bytes reader",
			imports: `"io"`,
			body: `body := http.MaxBytesReader(w, r.Body, 1024)
	_, _ = io.ReadAll(body)`,
			want: 0,
		},
		{
			name:    "external response body",
			imports: `"io"`,
			body: `res := &http.Response{Body: http.NoBody}
	_, _ = io.ReadAll(res.Body)`,
			want: 0,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			source := []byte(`package githubhttp

import (
	"context"
	"net/http"
	` + test.imports + `
)

func HandleWebhook(_ context.Context, w http.ResponseWriter, r *http.Request) error {
	` + test.body + `
	return nil
}
`)
			findings := inspectFixture(t, "internal/modules/github/http/webhook.go", source)
			assertRuleCount(t, findings, ruleUnsafeRawRequestBodyRead, test.want)
		})
	}
}

func TestModuleCycleFindingsAreDeterministicAndExcludeDAGs(t *testing.T) {
	t.Parallel()

	edges := []moduleImportEdge{
		{Source: "workspaces", Target: "subscriptions", Path: "internal/modules/workspaces/service/billing.go", Line: 9},
		{Source: "subscriptions", Target: "workspaces", Path: "internal/modules/subscriptions/service/workspaces.go", Line: 7},
		{Source: "stories", Target: "comments", Path: "internal/modules/stories/service/comments.go", Line: 8},
	}
	findings := findModuleCycleFindings(edges)
	assertRuleCount(t, findings, ruleModuleDependencyCycle, 2)
	for _, finding := range findings {
		if !strings.Contains(finding.Detail, "subscriptions -> workspaces -> subscriptions") &&
			!strings.Contains(finding.Detail, "workspaces -> subscriptions -> workspaces") {
			t.Errorf("cycle finding does not contain a concrete cycle: %s", finding.Detail)
		}
	}

	reversed := append([]moduleImportEdge(nil), edges...)
	for left, right := 0, len(reversed)-1; left < right; left, right = left+1, right-1 {
		reversed[left], reversed[right] = reversed[right], reversed[left]
	}
	second := findModuleCycleFindings(reversed)
	if len(second) != len(findings) {
		t.Fatalf("cycle scan changed finding count after input reorder: %d != %d", len(second), len(findings))
	}
	for index := range findings {
		if findings[index] != second[index] {
			t.Fatalf("cycle scan is nondeterministic at %d: %#v != %#v", index, findings[index], second[index])
		}
	}
}

func TestInspectModuleImportEdgesUsesProductionModuleImports(t *testing.T) {
	t.Parallel()

	source := []byte(`package subscriptionsservice

import (
	"github.com/complexus-tech/projects-api/internal/modules/subscriptions/repository"
	"github.com/complexus-tech/projects-api/internal/modules/workspaces/service"
	"github.com/complexus-tech/projects-api/internal/platform/auth"
)

var _ = repository.Repository{}
var _ = service.Service{}
var _ auth.Actor
`)
	edges, err := inspectModuleImportEdges(token.NewFileSet(), "internal/modules/subscriptions/service/service.go", source)
	if err != nil {
		t.Fatalf("inspect module imports: %v", err)
	}
	if len(edges) != 1 {
		t.Fatalf("inspect module imports produced %d edges, want 1: %#v", len(edges), edges)
	}
	if edges[0].Source != "subscriptions" || edges[0].Target != "workspaces" {
		t.Fatalf("unexpected module edge: %#v", edges[0])
	}
}
