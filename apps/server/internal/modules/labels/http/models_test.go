package labelshttp

import (
	"testing"

	labels "github.com/complexus-tech/projects-api/internal/modules/labels/service"
)

func TestToAppLabelsResponseReportsTheNextPageOnlyWhenPresent(t *testing.T) {
	t.Parallel()

	withMore := toAppLabelsResponse([]labels.CoreLabel{}, 3, 25, true)
	if withMore.Pagination.NextPage != 4 || !withMore.Pagination.HasMore {
		t.Fatalf("pagination with more results = %#v", withMore.Pagination)
	}
	complete := toAppLabelsResponse(nil, 3, 25, false)
	if complete.Pagination.NextPage != 0 || complete.Pagination.HasMore {
		t.Fatalf("complete pagination = %#v", complete.Pagination)
	}
}
