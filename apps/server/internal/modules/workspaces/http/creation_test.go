package workspaceshttp

import (
	"net/http/httptest"
	"strings"
	"testing"

	workspaces "github.com/complexus-tech/projects-api/internal/modules/workspaces/service"
	"github.com/complexus-tech/projects-api/pkg/web"
	"github.com/stretchr/testify/require"
)

func TestCreateWorkspaceDecodesOptionalExampleChoices(t *testing.T) {
	t.Parallel()
	include, exclude := true, false

	tests := []struct {
		name            string
		fields          string
		includeExamples *bool
		workType        workspaces.WorkType
		invalid         bool
	}{
		{name: "legacy request"},
		{name: "empty workspace", fields: `,"includeExamples":false`, includeExamples: &exclude},
		{name: "general examples", fields: `,"includeExamples":true`, includeExamples: &include},
		{name: "product", fields: `,"includeExamples":true,"workType":"product"`, includeExamples: &include, workType: workspaces.WorkTypeProduct},
		{name: "marketing", fields: `,"includeExamples":true,"workType":"marketing"`, includeExamples: &include, workType: workspaces.WorkTypeMarketing},
		{name: "operations", fields: `,"includeExamples":true,"workType":"operations"`, includeExamples: &include, workType: workspaces.WorkTypeOperations},
		{name: "personal", fields: `,"includeExamples":true,"workType":"personal"`, includeExamples: &include, workType: workspaces.WorkTypePersonal},
		{name: "general", fields: `,"includeExamples":true,"workType":"general"`, includeExamples: &include, workType: workspaces.WorkTypeGeneral},
		{name: "work type without examples flag", fields: `,"workType":"product"`, workType: workspaces.WorkTypeProduct},
		{name: "unknown work type", fields: `,"includeExamples":true,"workType":"unknown"`, invalid: true},
		{name: "unknown work type without examples", fields: `,"includeExamples":false,"workType":"unknown"`, invalid: true},
		{name: "examples must be boolean", fields: `,"includeExamples":"false"`, invalid: true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			request := httptest.NewRequest("POST", "/workspaces", strings.NewReader(
				`{"name":"New workspace","slug":"new-workspace","teamSize":"1-10"`+test.fields+`}`,
			))
			request.Header.Set("Content-Type", "application/json")
			var input AppNewWorkspace
			err := web.Decode(request, &input)
			if test.invalid {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.Equal(t, test.includeExamples, input.IncludeExamples)
			require.Equal(t, test.workType, input.WorkType)
			require.Equal(t, "New workspace", input.Name)
			require.Equal(t, "new-workspace", input.Slug)
			require.Equal(t, "1-10", input.TeamSize)
		})
	}
}
