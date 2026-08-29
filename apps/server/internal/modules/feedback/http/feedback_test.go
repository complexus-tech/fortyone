package feedbackhttp

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	feedback "github.com/complexus-tech/projects-api/internal/modules/feedback/service"
	teams "github.com/complexus-tech/projects-api/internal/modules/teams/service"
	"github.com/google/uuid"
)

func TestHTTPStatusClassifiesFeedbackErrors(t *testing.T) {
	t.Parallel()

	if status := httpStatus(fmt.Errorf("%w: invalid status", feedback.ErrInvalidInput)); status != http.StatusBadRequest {
		t.Fatalf("invalid input status = %d, want %d", status, http.StatusBadRequest)
	}
	if status := httpStatus(feedback.ErrStoryManaged); status != http.StatusConflict {
		t.Fatalf("story-managed status = %d, want %d", status, http.StatusConflict)
	}
	if status := httpStatus(feedback.ErrNotFound); status != http.StatusNotFound {
		t.Fatalf("not found status = %d, want %d", status, http.StatusNotFound)
	}
	if status := httpStatus(feedback.ErrParticipationNotAllowed); status != http.StatusForbidden {
		t.Fatalf("participation status = %d, want %d", status, http.StatusForbidden)
	}
	if status := httpStatus(feedback.ErrAuthenticationRequired); status != http.StatusUnauthorized {
		t.Fatalf("authentication status = %d, want %d", status, http.StatusUnauthorized)
	}
	if status := httpStatus(errors.New("database unavailable")); status != http.StatusInternalServerError {
		t.Fatalf("unexpected error status = %d, want %d", status, http.StatusInternalServerError)
	}
}

func TestParsePublicItemsQueryRejectsInvalidAuthorID(t *testing.T) {
	t.Parallel()

	for _, authorID := range []string{"not-a-uuid", uuid.Nil.String()} {
		t.Run(authorID, func(t *testing.T) {
			t.Parallel()

			request, err := http.NewRequest(http.MethodGet, "/?authorId="+authorID, nil)
			if err != nil {
				t.Fatalf("create request: %v", err)
			}
			_, err = parsePublicItemsQuery(request.URL.Query())
			if !errors.Is(err, feedback.ErrInvalidInput) {
				t.Fatalf("parse public items query error = %v, want invalid input", err)
			}
			if status := httpStatus(err); status != http.StatusBadRequest {
				t.Fatalf("invalid author status = %d, want %d", status, http.StatusBadRequest)
			}
		})
	}
}

func TestParsePublicItemsQueryRejectsInvalidBoardAndView(t *testing.T) {
	t.Parallel()

	for _, rawQuery := range []string{
		"boardId=not-a-uuid",
		"boardId=" + uuid.Nil.String(),
		"itemId=not-a-uuid",
		"itemId=" + uuid.Nil.String(),
		"view=full",
	} {
		request := httptest.NewRequest(http.MethodGet, "/?"+rawQuery, nil)

		_, err := parsePublicItemsQuery(request.URL.Query())

		if !errors.Is(err, feedback.ErrInvalidInput) {
			t.Fatalf("parsePublicItemsQuery(%q) error = %v, want invalid input", rawQuery, err)
		}
	}
}

func TestParsePublicItemsQueryParsesExactItemID(t *testing.T) {
	t.Parallel()
	itemID := uuid.New()
	request := httptest.NewRequest(http.MethodGet, "/?itemId="+itemID.String()+"&view=summary", nil)

	input, err := parsePublicItemsQuery(request.URL.Query())

	if err != nil {
		t.Fatalf("parsePublicItemsQuery error = %v", err)
	}
	if input.ItemID != itemID {
		t.Fatalf("item id = %s, want %s", input.ItemID, itemID)
	}
}

func TestPublicContributorIDRejectsInvalidValues(t *testing.T) {
	t.Parallel()

	for _, value := range []string{"not-a-uuid", uuid.Nil.String()} {
		request := httptest.NewRequest(http.MethodGet, "/", nil)
		request.SetPathValue("authorId", value)

		_, err := publicContributorID(request)

		if !errors.Is(err, feedback.ErrInvalidInput) {
			t.Fatalf("publicContributorID(%q) error = %v, want invalid input", value, err)
		}
	}
}

func TestParseContributorPaginationUsesExplicitValues(t *testing.T) {
	t.Parallel()

	request := httptest.NewRequest(http.MethodGet, "/?page=3&pageSize=25", nil)
	params, err := parseContributorPagination(request.URL.Query())
	if err != nil {
		t.Fatalf("parse contributor pagination: %v", err)
	}

	if params.Page != 3 || params.PageSize != 25 {
		t.Fatalf("pagination = %#v, want (3, 25)", params)
	}
}

func TestPublicFeedbackBotTrapRejectsFilledWebsite(t *testing.T) {
	t.Parallel()

	if err := validatePublicItemBotTrap(AppCreatePublicItem{}); err != nil {
		t.Fatalf("empty bot trap rejected: %v", err)
	}
	if err := validatePublicItemBotTrap(AppCreatePublicItem{Website: "https://spam.example"}); err == nil {
		t.Fatal("filled bot trap accepted")
	}
}

type teamAccessStub struct {
	err error
}

func (s teamAccessStub) GetByID(_ context.Context, teamID, workspaceID, _ uuid.UUID) (teams.CoreTeam, error) {
	if s.err != nil {
		return teams.CoreTeam{}, s.err
	}
	return teams.CoreTeam{ID: teamID, Workspace: workspaceID}, nil
}

func TestAuthorizeTeamMasksMembershipFailuresAsNotFound(t *testing.T) {
	t.Parallel()

	handler := New(nil, teamAccessStub{err: teams.ErrTeamNotFound}, nil, nil)
	err := handler.authorizeTeam(context.Background(), uuid.New(), uuid.New(), uuid.New())

	if !errors.Is(err, feedback.ErrNotFound) {
		t.Fatalf("authorize team error = %v, want not found", err)
	}
}

func TestAuthorizeTeamPreservesInfrastructureFailures(t *testing.T) {
	t.Parallel()

	infrastructureErr := errors.New("database unavailable")
	handler := New(nil, teamAccessStub{err: infrastructureErr}, nil, nil)
	err := handler.authorizeTeam(context.Background(), uuid.New(), uuid.New(), uuid.New())

	if !errors.Is(err, infrastructureErr) {
		t.Fatalf("authorize team error = %v, want infrastructure failure", err)
	}
}

func TestAppPortalDoesNotExposeRemovedDescription(t *testing.T) {
	t.Parallel()

	payload, err := json.Marshal(toAppPortal(feedback.CorePortal{ID: uuid.New()}))
	if err != nil {
		t.Fatalf("marshal portal: %v", err)
	}
	if strings.Contains(string(payload), "description") {
		t.Fatalf("portal payload unexpectedly contains description: %s", payload)
	}
}

func TestAppItemIncludesThirtyDayRestoreWindow(t *testing.T) {
	t.Parallel()

	deletedAt := time.Date(2026, time.July, 23, 8, 0, 0, 0, time.UTC)
	item := toAppItem(feedback.CoreItem{
		ID:        uuid.New(),
		DeletedAt: &deletedAt,
	}, nil, nil)

	if item.DeletedAt == nil || !item.DeletedAt.Equal(deletedAt) {
		t.Fatalf("deleted at = %v, want %v", item.DeletedAt, deletedAt)
	}
	wantRestoreUntil := deletedAt.Add(30 * 24 * time.Hour)
	if item.RestoreUntil == nil || !item.RestoreUntil.Equal(wantRestoreUntil) {
		t.Fatalf("restore until = %v, want %v", item.RestoreUntil, wantRestoreUntil)
	}
}

func TestAppItemUsesNullAuthorForAnonymousFeedback(t *testing.T) {
	t.Parallel()

	payload, err := json.Marshal(toAppItem(feedback.CoreItem{ID: uuid.New(), AuthorName: "Anonymous"}, nil, nil))
	if err != nil {
		t.Fatalf("marshal anonymous item: %v", err)
	}
	if !strings.Contains(string(payload), `"authorId":null`) {
		t.Fatalf("anonymous item authorId should be null: %s", payload)
	}
}

func TestAppBoardReviewerUsesAutoSaveContract(t *testing.T) {
	t.Parallel()

	userID := uuid.New()
	avatarURL := "https://cdn.example.com/ada.jpg"
	payload, err := json.Marshal(toAppBoardReviewer(feedback.CoreBoardReviewer{
		UserID:         userID,
		Name:           "Ada Lovelace",
		Email:          "ada@example.com",
		AvatarURL:      &avatarURL,
		Role:           "member",
		EmailFrequency: feedback.EmailFrequencyWeekly,
	}))
	if err != nil {
		t.Fatalf("marshal reviewer: %v", err)
	}

	var reviewer AppBoardReviewer
	if err := json.Unmarshal(payload, &reviewer); err != nil {
		t.Fatalf("unmarshal reviewer: %v", err)
	}
	if reviewer.UserID != userID || reviewer.EmailFrequency != feedback.EmailFrequencyWeekly {
		t.Fatalf("reviewer payload = %+v", reviewer)
	}
	if reviewer.AvatarURL == nil || *reviewer.AvatarURL != avatarURL {
		t.Fatalf("reviewer avatar = %v, want %q", reviewer.AvatarURL, avatarURL)
	}
}

type profileImageResolverStub struct {
	resolved map[string]string
	errors   map[string]error
	calls    map[string]int
}

func (s *profileImageResolverStub) ResolveProfileImageURL(_ context.Context, avatar string, _ time.Duration) (string, error) {
	if s.calls == nil {
		s.calls = make(map[string]int)
	}
	s.calls[avatar]++
	if err := s.errors[avatar]; err != nil {
		return "", err
	}
	return s.resolved[avatar], nil
}

func stringPointer(value string) *string {
	return &value
}

func TestResolvePortalAvatarsUsesPresignedURLs(t *testing.T) {
	t.Parallel()

	itemID := uuid.New()
	resolver := &profileImageResolverStub{
		resolved: map[string]string{
			"profiles/joseph.webp": "https://signed.example.com/profiles/joseph.webp",
		},
	}
	handler := New(nil, nil, resolver, nil)
	portal := feedback.CorePortalSnapshot{
		Items: []feedback.CoreItem{
			{ID: itemID, AuthorAvatar: stringPointer("profiles/joseph.webp")},
		},
		Comments: []feedback.CoreComment{
			{ItemID: itemID, AuthorAvatar: stringPointer("profiles/joseph.webp")},
		},
	}

	handler.resolvePortalAvatars(context.Background(), &portal)

	want := "https://signed.example.com/profiles/joseph.webp"
	if portal.Items[0].AuthorAvatar == nil || *portal.Items[0].AuthorAvatar != want {
		t.Fatalf("item avatar = %v, want %q", portal.Items[0].AuthorAvatar, want)
	}
	if portal.Comments[0].AuthorAvatar == nil || *portal.Comments[0].AuthorAvatar != want {
		t.Fatalf("comment avatar = %v, want %q", portal.Comments[0].AuthorAvatar, want)
	}
	if got := resolver.calls["profiles/joseph.webp"]; got != 1 {
		t.Fatalf("resolver calls = %d, want 1", got)
	}
}

func TestResolveSimilarItemAvatarsUsesPresignedURLs(t *testing.T) {
	t.Parallel()

	resolver := &profileImageResolverStub{
		resolved: map[string]string{
			"profiles/joseph.webp": "https://signed.example.com/profiles/joseph.webp",
		},
	}
	handler := New(nil, nil, resolver, nil)
	items := []feedback.CoreSimilarItem{
		{AuthorAvatar: stringPointer("profiles/joseph.webp")},
		{AuthorAvatar: stringPointer("profiles/joseph.webp")},
	}

	handler.resolveSimilarItemAvatars(context.Background(), items)

	want := "https://signed.example.com/profiles/joseph.webp"
	for i, item := range items {
		if item.AuthorAvatar == nil || *item.AuthorAvatar != want {
			t.Fatalf("item %d avatar = %v, want %q", i, item.AuthorAvatar, want)
		}
	}
	if got := resolver.calls["profiles/joseph.webp"]; got != 1 {
		t.Fatalf("resolver calls = %d, want 1", got)
	}
}

func TestRespondPublicContributorUsesPresignedAvatarAndDoesNotExposeEmail(t *testing.T) {
	t.Parallel()

	resolver := &profileImageResolverStub{
		resolved: map[string]string{
			"profiles/ada.webp": "https://signed.example.com/profiles/ada.webp",
		},
	}
	handler := New(nil, nil, resolver, nil)
	recorder := httptest.NewRecorder()
	avatar := "profiles/ada.webp"

	err := handler.respondPublicContributor(context.Background(), recorder, feedback.CoreContributor{
		ID:        uuid.New(),
		Name:      "Ada Lovelace",
		AvatarURL: &avatar,
		JoinedAt:  time.Now(),
		Stats: feedback.CoreContributorStats{
			FeedbackCount: 2,
			CommentCount:  3,
			VoteScore:     -1,
		},
	})

	if err != nil {
		t.Fatalf("respond contributor: %v", err)
	}
	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", recorder.Code)
	}
	body := recorder.Body.String()
	if !strings.Contains(body, "https://signed.example.com/profiles/ada.webp") {
		t.Fatalf("response does not contain presigned avatar: %s", body)
	}
	if strings.Contains(strings.ToLower(body), "email") {
		t.Fatalf("response unexpectedly exposes an email field: %s", body)
	}
	if !strings.Contains(body, `"voteScore":-1`) {
		t.Fatalf("response does not contain net vote score: %s", body)
	}
}

func TestRespondPublicContributorCommentsBuildsStablePagination(t *testing.T) {
	t.Parallel()

	recorder := httptest.NewRecorder()
	err := respondPublicContributorComments(context.Background(), recorder, feedback.CoreContributorCommentsPage{
		Comments: []feedback.CoreContributorComment{{
			ID:            uuid.New(),
			ItemID:        uuid.New(),
			FeedbackTitle: "Repair the crossing",
			FeedbackSlug:  "repair-the-crossing",
			Body:          "This would make the crossing safer.",
		}},
		Page:     2,
		PageSize: 20,
		HasMore:  true,
	})

	if err != nil {
		t.Fatalf("respond contributor comments: %v", err)
	}
	body := recorder.Body.String()
	if !strings.Contains(body, `"feedback":{"id":`) || !strings.Contains(body, `"slug":"repair-the-crossing"`) {
		t.Fatalf("response does not contain feedback context: %s", body)
	}
	if !strings.Contains(body, `"page":2`) || !strings.Contains(body, `"nextPage":3`) {
		t.Fatalf("response has incorrect pagination: %s", body)
	}
}

func TestResolvePortalAvatarsOmitsUnresolvableAvatar(t *testing.T) {
	t.Parallel()

	resolver := &profileImageResolverStub{
		errors: map[string]error{
			"missing.webp": errors.New("presign failed"),
		},
	}
	handler := New(nil, nil, resolver, nil)
	portal := feedback.CorePortalSnapshot{
		Items: []feedback.CoreItem{
			{ID: uuid.New(), AuthorAvatar: stringPointer("missing.webp")},
		},
	}

	handler.resolvePortalAvatars(context.Background(), &portal)

	if portal.Items[0].AuthorAvatar != nil {
		t.Fatalf("item avatar = %q, want nil", *portal.Items[0].AuthorAvatar)
	}
}

func TestResolvePortalAvatarsOnlyResolvesVisibleComments(t *testing.T) {
	t.Parallel()

	resolver := &profileImageResolverStub{
		resolved: map[string]string{
			"visible.webp": "https://signed.example.com/visible.webp",
		},
	}
	handler := New(nil, nil, resolver, nil)
	visibleItemID := uuid.New()
	portal := feedback.CorePortalSnapshot{
		Items: []feedback.CoreItem{{ID: visibleItemID}},
		Comments: []feedback.CoreComment{
			{ItemID: visibleItemID, AuthorAvatar: stringPointer("visible.webp")},
			{ItemID: uuid.New(), AuthorAvatar: stringPointer("not-visible.webp")},
		},
	}

	handler.resolvePortalAvatars(context.Background(), &portal)

	if got := resolver.calls["visible.webp"]; got != 1 {
		t.Fatalf("visible avatar resolver calls = %d, want 1", got)
	}
	if got := resolver.calls["not-visible.webp"]; got != 0 {
		t.Fatalf("hidden avatar resolver calls = %d, want 0", got)
	}
}
