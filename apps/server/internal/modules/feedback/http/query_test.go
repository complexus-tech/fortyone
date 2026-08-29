package feedbackhttp

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	feedback "github.com/complexus-tech/projects-api/internal/modules/feedback/service"
	"github.com/complexus-tech/projects-api/pkg/web"
	"github.com/google/uuid"
)

func TestParsePublicItemsQueryProducesTypedBoundedInput(t *testing.T) {
	t.Parallel()

	boardID, itemID, authorID := uuid.New(), uuid.New(), uuid.New()
	input, err := parsePublicItemsQuery(url.Values{
		"boardId":  {boardID.String()},
		"itemId":   {itemID.String()},
		"authorId": {authorID.String()},
		"status":   {feedback.StatusPlanned},
		"search":   {"  dark mode  "},
		"sort":     {"newest"},
		"view":     {"summary"},
		"page":     {"2"},
		"pageSize": {"500"},
	})
	if err != nil {
		t.Fatalf("parse public items query: %v", err)
	}
	if input.BoardID == nil || *input.BoardID != boardID || input.ItemID != itemID || input.AuthorID != authorID {
		t.Fatalf("identity filters = %#v", input)
	}
	if input.Status != feedback.StatusPlanned || input.Search != "dark mode" || input.Sort != "newest" || !input.SummaryOnly {
		t.Fatalf("content filters = %#v", input)
	}
	if input.Page != 2 || input.PageSize != maximumFeedbackPageSize {
		t.Fatalf("pagination = (%d, %d)", input.Page, input.PageSize)
	}

	defaults, err := parsePublicItemsQuery(url.Values{})
	if err != nil {
		t.Fatalf("parse default public items query: %v", err)
	}
	if defaults.Page != 1 || defaults.PageSize != defaultPublicItemsPageSize {
		t.Fatalf("default pagination = (%d, %d)", defaults.Page, defaults.PageSize)
	}
}

func TestFeedbackQueryParsersPreserveEndpointDefaultsAndCaps(t *testing.T) {
	t.Parallel()

	activity, err := parseContributorActivityQuery(url.Values{"type": {"comment"}, "pageSize": {"500"}})
	if err != nil {
		t.Fatalf("parse contributor activity query: %v", err)
	}
	if activity.ActivityType != "comment" || activity.Pagination.Page != 1 || activity.Pagination.PageSize != maximumFeedbackPageSize {
		t.Fatalf("activity query = %#v", activity)
	}

	updates, err := parseUpdatesPagination(url.Values{"pageSize": {"500"}})
	if err != nil {
		t.Fatalf("parse updates pagination: %v", err)
	}
	if updates.Page != 1 || updates.PageSize != maximumFeedbackPageSize {
		t.Fatalf("updates pagination = %#v", updates)
	}

	team, err := parseTeamFeedbackListQuery(url.Values{})
	if err != nil {
		t.Fatalf("parse team feedback query: %v", err)
	}
	if team.Status != "active" || team.Pagination.Page != 1 || team.Pagination.PageSize != defaultTeamFeedbackPageSize {
		t.Fatalf("team query = %#v", team)
	}

	similar, err := parsePublicSimilarItemsQuery(url.Values{
		"title": {"  Dark mode  "}, "description": {"  Please add it  "}, "limit": {"500"},
	})
	if err != nil {
		t.Fatalf("parse similar items query: %v", err)
	}
	if similar.Title != "Dark mode" || similar.Description != "Please add it" || similar.Limit != maximumSimilarItemsLimit {
		t.Fatalf("similar query = %#v", similar)
	}
	similarDefaults, err := parsePublicSimilarItemsQuery(url.Values{})
	if err != nil || similarDefaults.Limit != defaultSimilarItemsLimit {
		t.Fatalf("default similar query = %#v, %v", similarDefaults, err)
	}

	candidates, err := parseCandidateItemsQuery(url.Values{"limit": {"500"}})
	if err != nil {
		t.Fatalf("parse candidate query: %v", err)
	}
	if candidates.Limit != maximumCandidateItemsLimit || candidates.Search != "" {
		t.Fatalf("candidate query = %#v", candidates)
	}
	candidateDefaults, err := parseCandidateItemsQuery(url.Values{})
	if err != nil || candidateDefaults.Limit != defaultCandidateItemsLimit {
		t.Fatalf("default candidate query = %#v, %v", candidateDefaults, err)
	}
}

func TestFeedbackQueryParsersRejectUnsafeInputWithoutEchoingIt(t *testing.T) {
	t.Parallel()

	parsers := map[string]struct {
		query url.Values
		cause error
		parse func(url.Values) error
	}{
		"repeated page": {
			query: url.Values{"page": {"1", "2"}}, cause: web.ErrRepeatedQueryParameter,
			parse: publicItemsQueryError,
		},
		"malformed page": {
			query: url.Values{"page": {"sensitive-page-value"}}, cause: web.ErrInvalidQueryParameter,
			parse: publicItemsQueryError,
		},
		"integer overflow": {
			query: url.Values{"page": {strings.Repeat("9", maximumFeedbackIntegerBytes)}}, cause: web.ErrInvalidQueryParameter,
			parse: publicItemsQueryError,
		},
		"oversized integer": {
			query: url.Values{"page": {strings.Repeat("9", maximumFeedbackIntegerBytes+1)}}, cause: web.ErrQueryParameterTooLong,
			parse: publicItemsQueryError,
		},
		"offset overflow": {
			query: url.Values{"page": {"2147483649"}, "pageSize": {"1"}}, cause: web.ErrInvalidQueryParameter,
			parse: publicItemsQueryError,
		},
		"repeated board": {
			query: url.Values{"boardId": {uuid.NewString(), uuid.NewString()}}, cause: web.ErrRepeatedQueryParameter,
			parse: publicItemsQueryError,
		},
		"invalid item": {
			query: url.Values{"itemId": {"sensitive-item-value"}}, cause: web.ErrInvalidQueryParameter,
			parse: publicItemsQueryError,
		},
		"oversized author": {
			query: url.Values{"authorId": {strings.Repeat("x", web.DefaultMaxQueryParameterBytes+1)}}, cause: web.ErrQueryParameterTooLong,
			parse: publicItemsQueryError,
		},
		"repeated search": {
			query: url.Values{"search": {"first", "second"}}, cause: web.ErrRepeatedQueryParameter,
			parse: publicItemsQueryError,
		},
		"oversized search": {
			query: url.Values{"search": {strings.Repeat("x", maximumFeedbackSearchRunes*4+1)}}, cause: web.ErrQueryParameterTooLong,
			parse: publicItemsQueryError,
		},
		"too many search characters": {
			query: url.Values{"search": {strings.Repeat("é", maximumFeedbackSearchRunes+1)}}, cause: web.ErrInvalidQueryParameter,
			parse: publicItemsQueryError,
		},
		"invalid search encoding": {
			query: url.Values{"search": {string([]byte{0xff})}}, cause: web.ErrInvalidQueryParameter,
			parse: publicItemsQueryError,
		},
		"invalid status": {
			query: url.Values{"status": {"sensitive-status-value"}}, cause: web.ErrInvalidQueryParameter,
			parse: publicItemsQueryError,
		},
		"invalid sort": {
			query: url.Values{"sort": {"sensitive-sort-value"}}, cause: web.ErrInvalidQueryParameter,
			parse: publicItemsQueryError,
		},
		"invalid view": {
			query: url.Values{"view": {"sensitive-view-value"}}, cause: web.ErrInvalidQueryParameter,
			parse: publicItemsQueryError,
		},
		"repeated title": {
			query: url.Values{"title": {"first", "second"}}, cause: web.ErrRepeatedQueryParameter,
			parse: publicSimilarItemsQueryError,
		},
		"oversized title": {
			query: url.Values{"title": {strings.Repeat("x", maximumFeedbackTitleRunes*4+1)}}, cause: web.ErrQueryParameterTooLong,
			parse: publicSimilarItemsQueryError,
		},
		"too many description characters": {
			query: url.Values{"description": {strings.Repeat("é", maximumFeedbackDescriptionRunes+1)}}, cause: web.ErrInvalidQueryParameter,
			parse: publicSimilarItemsQueryError,
		},
		"invalid similar limit": {
			query: url.Values{"limit": {"limit-secret"}}, cause: web.ErrInvalidQueryParameter,
			parse: publicSimilarItemsQueryError,
		},
		"invalid activity type": {
			query: url.Values{"type": {"sensitive-type-value"}}, cause: web.ErrInvalidQueryParameter,
			parse: contributorActivityQueryError,
		},
		"repeated activity type": {
			query: url.Values{"type": {"feedback", "comment"}}, cause: web.ErrRepeatedQueryParameter,
			parse: contributorActivityQueryError,
		},
		"invalid team status": {
			query: url.Values{"status": {"sensitive-team-status"}}, cause: web.ErrInvalidQueryParameter,
			parse: teamFeedbackQueryError,
		},
		"repeated team search": {
			query: url.Values{"search": {"first", "second"}}, cause: web.ErrRepeatedQueryParameter,
			parse: teamFeedbackQueryError,
		},
		"invalid candidate limit": {
			query: url.Values{"limit": {"0"}}, cause: web.ErrInvalidQueryParameter,
			parse: candidateItemsQueryError,
		},
		"repeated updates page": {
			query: url.Values{"page": {"1", "2"}}, cause: web.ErrRepeatedQueryParameter,
			parse: updatesPaginationError,
		},
	}

	for name, test := range parsers {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			err := test.parse(test.query)
			if !errors.Is(err, feedback.ErrInvalidInput) || !errors.Is(err, test.cause) {
				t.Fatalf("error = %v, want invalid input and %v", err, test.cause)
			}
			for _, values := range test.query {
				for _, value := range values {
					if value != "" && value != "feedback" && strings.Contains(err.Error(), value) {
						t.Fatalf("error %q exposes query value", err)
					}
				}
			}
		})
	}
}

func TestPublicQueryHandlersReturnSafeBadRequestsBeforeServiceCalls(t *testing.T) {
	t.Parallel()

	for name, test := range map[string]struct {
		target  string
		handler web.Handler
		secret  string
	}{
		"portal": {
			target:  "/portals/planning/feedback?search=sensitive-search&search=second",
			handler: New(nil, nil, nil, nil).GetPortal,
			secret:  "sensitive-search",
		},
		"similar": {
			target:  "/portals/planning/feedback/similar?limit=sensitive-limit&limit=3",
			handler: New(nil, nil, nil, nil).ListPublicSimilarItems,
			secret:  "sensitive-limit",
		},
		"updates": {
			target:  "/portals/planning/feedback/updates?page=sensitive-page&page=2",
			handler: New(nil, nil, nil, nil).ListPublicUpdates,
			secret:  "sensitive-page",
		},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			request := httptest.NewRequest(http.MethodGet, test.target, nil)
			request.SetPathValue("portalSlug", "planning")
			recorder := httptest.NewRecorder()
			if err := test.handler(context.Background(), recorder, request); err != nil {
				t.Fatalf("handle request: %v", err)
			}
			if recorder.Code != http.StatusBadRequest {
				t.Fatalf("status = %d, want %d", recorder.Code, http.StatusBadRequest)
			}
			if strings.Contains(recorder.Body.String(), test.secret) {
				t.Fatalf("response exposes query value: %s", recorder.Body.String())
			}
		})
	}
}

func publicItemsQueryError(values url.Values) error {
	_, err := parsePublicItemsQuery(values)
	return err
}

func publicSimilarItemsQueryError(values url.Values) error {
	_, err := parsePublicSimilarItemsQuery(values)
	return err
}

func contributorActivityQueryError(values url.Values) error {
	_, err := parseContributorActivityQuery(values)
	return err
}

func teamFeedbackQueryError(values url.Values) error {
	_, err := parseTeamFeedbackListQuery(values)
	return err
}

func candidateItemsQueryError(values url.Values) error {
	_, err := parseCandidateItemsQuery(values)
	return err
}

func updatesPaginationError(values url.Values) error {
	_, err := parseUpdatesPagination(values)
	return err
}
