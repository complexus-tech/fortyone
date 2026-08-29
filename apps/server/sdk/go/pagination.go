package fortyone

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	openapi_types "github.com/oapi-codegen/runtime/types"
)

var ErrInvalidPaginationCursor = errors.New("FortyOne API returned an invalid or repeated pagination cursor")

// StoryPageClient is the narrow generated operation required by StoryPager.
// It makes the pager straightforward to test and to wrap with observability.
type StoryPageClient interface {
	ListStoriesWithResponse(context.Context, ComponentsCommonWorkspaceId, *ListStoriesParams, ...RequestEditorFn) (*ListStoriesResponse, error)
}

// StoryPaginationOptions is copied into a pager. Limit must be from 1 through
// 100 when provided. Cursor values remain opaque to the SDK.
type StoryPaginationOptions struct {
	TeamID *openapi_types.UUID
	Limit  int
}

// StoryPager retrieves deterministic pages. A pager is stateful and must not
// be used concurrently; create one pager per traversal.
type StoryPager struct {
	client      StoryPageClient
	workspaceID ComponentsCommonWorkspaceId
	teamID      *openapi_types.UUID
	limit       *int
	cursor      *string
	seen        map[string]struct{}
	done        bool
}

func NewStoryPager(client StoryPageClient, workspaceID ComponentsCommonWorkspaceId, options StoryPaginationOptions) (*StoryPager, error) {
	if client == nil {
		return nil, errors.New("story pager requires a client")
	}
	var limit *int
	if options.Limit != 0 {
		if options.Limit < 1 || options.Limit > 100 {
			return nil, errors.New("story page limit must be from 1 through 100")
		}
		value := options.Limit
		limit = &value
	}
	return &StoryPager{
		client:      client,
		workspaceID: workspaceID,
		teamID:      options.TeamID,
		limit:       limit,
		seen:        make(map[string]struct{}),
	}, nil
}

// NextPage returns the next page and whether a page was available.
func (pager *StoryPager) NextPage(ctx context.Context) (StoryPage, bool, error) {
	if pager.done {
		return StoryPage{}, false, nil
	}
	response, err := pager.client.ListStoriesWithResponse(ctx, pager.workspaceID, &ListStoriesParams{
		Cursor: pager.cursor,
		Limit:  pager.limit,
		TeamId: pager.teamID,
	})
	if err != nil {
		return StoryPage{}, false, fmt.Errorf("list FortyOne stories: %w", err)
	}
	if response.JSON200 == nil {
		var headers http.Header
		if response.HTTPResponse != nil {
			headers = response.HTTPResponse.Header
		}
		return StoryPage{}, false, NewAPIError(response.StatusCode(), headers, response.Body)
	}
	page := *response.JSON200
	if !page.Meta.HasMore {
		pager.done = true
		return page, true, nil
	}
	if page.Meta.NextCursor == nil || *page.Meta.NextCursor == "" {
		return StoryPage{}, false, ErrInvalidPaginationCursor
	}
	next := *page.Meta.NextCursor
	if _, exists := pager.seen[next]; exists {
		return StoryPage{}, false, ErrInvalidPaginationCursor
	}
	pager.seen[next] = struct{}{}
	pager.cursor = &next
	return page, true, nil
}
