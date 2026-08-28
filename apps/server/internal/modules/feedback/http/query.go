package feedbackhttp

import (
	"fmt"
	"math"
	"net/url"
	"strings"
	"unicode/utf8"

	feedback "github.com/complexus-tech/projects-api/internal/modules/feedback/service"
	"github.com/complexus-tech/projects-api/internal/platform/pagination"
	"github.com/complexus-tech/projects-api/pkg/web"
)

const (
	defaultPublicItemsPageSize      = 20
	defaultContributorPageSize      = 20
	defaultUpdatesPageSize          = 20
	defaultSimilarItemsLimit        = 3
	defaultCandidateItemsLimit      = 30
	maximumFeedbackPageSize         = 50
	maximumSimilarItemsLimit        = 5
	maximumCandidateItemsLimit      = 50
	maximumFeedbackOffset           = math.MaxInt32
	maximumFeedbackIntegerBytes     = 20
	maximumFeedbackSearchRunes      = 200
	maximumFeedbackTitleRunes       = 200
	maximumFeedbackDescriptionRunes = 20_000
	maximumFeedbackEnumRunes        = 32
)

type publicSimilarItemsQuery struct {
	Title       string
	Description string
	Limit       int
}

type contributorActivityQuery struct {
	ActivityType string
	Pagination   pagination.OffsetParams
}

type teamFeedbackListQuery struct {
	Status     string
	Search     string
	Pagination pagination.OffsetParams
}

type candidateItemsQuery struct {
	Search string
	Limit  int
}

func parsePublicItemsQuery(values url.Values) (feedback.CorePortalSnapshotInput, error) {
	params, err := parseFeedbackPagination(values, defaultPublicItemsPageSize)
	if err != nil {
		return feedback.CorePortalSnapshotInput{}, err
	}
	status, err := parseFeedbackString(values, "status", maximumFeedbackEnumRunes)
	if err != nil {
		return feedback.CorePortalSnapshotInput{}, err
	}
	if !validFeedbackListStatus(status, false) {
		return feedback.CorePortalSnapshotInput{}, invalidFeedbackQueryParameter("status")
	}
	search, err := parseFeedbackString(values, "search", maximumFeedbackSearchRunes)
	if err != nil {
		return feedback.CorePortalSnapshotInput{}, err
	}
	sort, err := parseFeedbackString(values, "sort", maximumFeedbackEnumRunes)
	if err != nil {
		return feedback.CorePortalSnapshotInput{}, err
	}
	if sort != "" && sort != "top" && sort != "newest" && sort != "oldest" {
		return feedback.CorePortalSnapshotInput{}, invalidFeedbackQueryParameter("sort")
	}
	view, err := parseFeedbackString(values, "view", maximumFeedbackEnumRunes)
	if err != nil {
		return feedback.CorePortalSnapshotInput{}, err
	}
	if view != "" && view != "summary" {
		return feedback.CorePortalSnapshotInput{}, invalidFeedbackQueryParameter("view")
	}

	boardID, err := web.OptionalUUIDQueryParameter(values, "boardId")
	if err != nil {
		return feedback.CorePortalSnapshotInput{}, feedbackQueryError(err)
	}
	itemID, err := web.OptionalUUIDQueryParameter(values, "itemId")
	if err != nil {
		return feedback.CorePortalSnapshotInput{}, feedbackQueryError(err)
	}
	authorID, err := web.OptionalUUIDQueryParameter(values, "authorId")
	if err != nil {
		return feedback.CorePortalSnapshotInput{}, feedbackQueryError(err)
	}

	input := feedback.CorePortalSnapshotInput{
		Status: status, BoardID: boardID, Search: search, Sort: sort,
		Page: params.Page, PageSize: params.PageSize, SummaryOnly: view == "summary",
	}
	if itemID != nil {
		input.ItemID = *itemID
	}
	if authorID != nil {
		input.AuthorID = *authorID
	}
	return input, nil
}

func parsePublicSimilarItemsQuery(values url.Values) (publicSimilarItemsQuery, error) {
	title, err := parseFeedbackString(values, "title", maximumFeedbackTitleRunes)
	if err != nil {
		return publicSimilarItemsQuery{}, err
	}
	description, err := parseFeedbackString(values, "description", maximumFeedbackDescriptionRunes)
	if err != nil {
		return publicSimilarItemsQuery{}, err
	}
	limit, err := parseFeedbackLimit(values, "limit", defaultSimilarItemsLimit, maximumSimilarItemsLimit)
	if err != nil {
		return publicSimilarItemsQuery{}, err
	}
	return publicSimilarItemsQuery{Title: title, Description: description, Limit: limit}, nil
}

func parseContributorActivityQuery(values url.Values) (contributorActivityQuery, error) {
	params, err := parseFeedbackPagination(values, defaultContributorPageSize)
	if err != nil {
		return contributorActivityQuery{}, err
	}
	activityType, err := parseFeedbackString(values, "type", maximumFeedbackEnumRunes)
	if err != nil {
		return contributorActivityQuery{}, err
	}
	if activityType != "" && activityType != "feedback" && activityType != "comment" {
		return contributorActivityQuery{}, invalidFeedbackQueryParameter("type")
	}
	return contributorActivityQuery{ActivityType: activityType, Pagination: params}, nil
}

func parseContributorPagination(values url.Values) (pagination.OffsetParams, error) {
	return parseFeedbackPagination(values, defaultContributorPageSize)
}

func parseUpdatesPagination(values url.Values) (pagination.OffsetParams, error) {
	return parseFeedbackPagination(values, defaultUpdatesPageSize)
}

func parseTeamFeedbackListQuery(values url.Values) (teamFeedbackListQuery, error) {
	params, err := parseFeedbackPagination(values, defaultTeamFeedbackPageSize)
	if err != nil {
		return teamFeedbackListQuery{}, err
	}
	status, err := parseFeedbackString(values, "status", maximumFeedbackEnumRunes)
	if err != nil {
		return teamFeedbackListQuery{}, err
	}
	if status == "" {
		status = "active"
	}
	if !validFeedbackListStatus(status, true) {
		return teamFeedbackListQuery{}, invalidFeedbackQueryParameter("status")
	}
	search, err := parseFeedbackString(values, "search", maximumFeedbackSearchRunes)
	if err != nil {
		return teamFeedbackListQuery{}, err
	}
	return teamFeedbackListQuery{Status: status, Search: search, Pagination: params}, nil
}

func parseCandidateItemsQuery(values url.Values) (candidateItemsQuery, error) {
	search, err := parseFeedbackString(values, "search", maximumFeedbackSearchRunes)
	if err != nil {
		return candidateItemsQuery{}, err
	}
	limit, err := parseFeedbackLimit(values, "limit", defaultCandidateItemsLimit, maximumCandidateItemsLimit)
	if err != nil {
		return candidateItemsQuery{}, err
	}
	return candidateItemsQuery{Search: search, Limit: limit}, nil
}

func parseFeedbackPagination(values url.Values, defaultPageSize int) (pagination.OffsetParams, error) {
	params, err := pagination.ParseOffsetQuery(values, pagination.OffsetQueryConfig{
		DefaultPageSize: defaultPageSize,
		MaximumPageSize: maximumFeedbackPageSize,
		MaximumOffset:   maximumFeedbackOffset,
	})
	if err != nil {
		return pagination.OffsetParams{}, feedbackQueryError(err)
	}
	return params, nil
}

func parseFeedbackLimit(values url.Values, name string, defaultValue, maximumValue int) (int, error) {
	value, present, err := web.OptionalIntegerQueryParameter(
		values, name, maximumFeedbackIntegerBytes, 1, math.MaxInt,
	)
	if err != nil {
		return 0, feedbackQueryError(err)
	}
	if !present {
		return defaultValue, nil
	}
	if value > maximumValue {
		value = maximumValue
	}
	return value, nil
}

func parseFeedbackString(values url.Values, name string, maximumRunes int) (string, error) {
	value, _, err := web.OptionalQueryParameter(values, name, maximumRunes*utf8.UTFMax)
	if err != nil {
		return "", feedbackQueryError(err)
	}
	if !utf8.ValidString(value) || strings.ContainsRune(value, '\x00') || utf8.RuneCountInString(value) > maximumRunes {
		return "", invalidFeedbackQueryParameter(name)
	}
	return strings.TrimSpace(value), nil
}

func validFeedbackListStatus(status string, allowTrashed bool) bool {
	switch status {
	case "", "active", "all":
		return true
	case feedback.ListStatusTrashed:
		return allowTrashed
	case feedback.StatusPending, feedback.StatusReviewing, feedback.StatusPlanned,
		feedback.StatusInProgress, feedback.StatusCompleted, feedback.StatusClosed:
		return true
	default:
		return false
	}
}

func invalidFeedbackQueryParameter(name string) error {
	return feedbackQueryError(&web.QueryParameterError{Name: name, Cause: web.ErrInvalidQueryParameter})
}

func feedbackQueryError(cause error) error {
	return fmt.Errorf("%w: %w", feedback.ErrInvalidInput, cause)
}
