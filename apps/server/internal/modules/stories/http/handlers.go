package storieshttp

// Package storieshttp provides bounded HTTP adapters for story use cases.

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	attachments "github.com/complexus-tech/projects-api/internal/modules/attachments/service"
	links "github.com/complexus-tech/projects-api/internal/modules/links/service"
	stories "github.com/complexus-tech/projects-api/internal/modules/stories/service"
	users "github.com/complexus-tech/projects-api/internal/modules/users/service"
	"github.com/complexus-tech/projects-api/pkg/cache"
	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/complexus-tech/projects-api/pkg/web"
	"github.com/google/uuid"
)

var (
	// ErrInvalidStoryID is returned when a story ID is not in proper UUID format.
	ErrInvalidStoryID = errors.New("story id is not in its proper form")
)

func storyReadStatus(err error) int {
	switch {
	case errors.Is(err, stories.ErrNotFound):
		return http.StatusNotFound
	case errors.Is(err, stories.ErrStoryReadForbidden):
		return http.StatusForbidden
	case errors.Is(err, stories.ErrDeleteForbidden):
		return http.StatusForbidden
	case errors.Is(err, stories.ErrInvalidStoryReference):
		return http.StatusBadRequest
	case errors.Is(err, stories.ErrInvalidStoryReadQuery):
		return http.StatusBadRequest
	default:
		return http.StatusInternalServerError
	}
}

func storyAttachmentStatus(err error) int {
	switch {
	case errors.Is(err, attachments.ErrNotFound):
		return http.StatusNotFound
	case errors.Is(err, attachments.ErrUnauthorized):
		return http.StatusForbidden
	case errors.Is(err, attachments.ErrInvalidFile):
		return http.StatusBadRequest
	default:
		return http.StatusInternalServerError
	}
}

func bulkDeleteStatus(err error) int {
	switch {
	case errors.Is(err, stories.ErrDeleteForbidden):
		return http.StatusForbidden
	case errors.Is(err, stories.ErrNotFound):
		return http.StatusNotFound
	default:
		return http.StatusInternalServerError
	}
}

func storyMutationStatus(err error) int {
	switch {
	case errors.Is(err, stories.ErrAutoSchedulingUnavailable):
		return http.StatusPaymentRequired
	case errors.Is(err, stories.ErrAutoSchedulingAccessCheckFailed):
		return http.StatusServiceUnavailable
	case errors.Is(err, stories.ErrNotFound):
		return http.StatusNotFound
	case errors.Is(err, stories.ErrDeleteForbidden),
		errors.Is(err, stories.ErrStoryMutationForbidden):
		return http.StatusForbidden
	case errors.Is(err, stories.ErrStoryChanged),
		errors.Is(err, stories.ErrAutoSchedulingOwnerLocked),
		errors.Is(err, stories.ErrAutoSchedulingLockEmpty):
		return http.StatusConflict
	case errors.Is(err, stories.ErrMayaAssignmentRequiresScheduling),
		errors.Is(err, stories.ErrMayaAssignmentRequiresDuration),
		errors.Is(err, stories.ErrMayaAssignmentRequiresDeliveryDate),
		errors.Is(err, stories.ErrInvalidStoryReference),
		errors.Is(err, stories.ErrInvalidStoryMediaReference),
		errors.Is(err, stories.ErrObjectiveKeyResultMismatch),
		errors.Is(err, stories.ErrInvalidStoryLabels),
		errors.Is(err, stories.ErrInvalidStoryMutation),
		errors.Is(err, stories.ErrInvalidEstimatedDuration),
		errors.Is(err, stories.ErrInvalidMinimumFocusBlock),
		errors.Is(err, stories.ErrEstimatedDurationTooLarge),
		errors.Is(err, stories.ErrMinimumFocusBlockTooLarge),
		errors.Is(err, stories.ErrFocusBlockRequiresDuration),
		errors.Is(err, stories.ErrFocusBlockExceedsDuration),
		errors.Is(err, stories.ErrInvalidAutoSchedulingStatus),
		errors.Is(err, stories.ErrLockedAutoSchedulingOff):
		return http.StatusBadRequest
	default:
		return http.StatusInternalServerError
	}
}

// Handlers provides HTTP handlers for story operations.
type Handlers struct {
	stories     *stories.Service
	users       storyUserReader
	links       *links.Service
	attachments *attachments.Service
	storyMedia  storyMediaService
	cache       *storyCache
	log         *logger.Logger
}

// storyCache preserves the mutation handlers' best-effort cache semantics
// after their database write has committed, while ensuring an invalidation
// failure is observable. Returning an HTTP failure at that point would invite a
// retry of an already-completed mutation.
type storyCache struct {
	*cache.Service
	log *logger.Logger
}

func newStoryCache(service *cache.Service, log *logger.Logger) *storyCache {
	if service == nil {
		return nil
	}
	return &storyCache{Service: service, log: log}
}

func (c *storyCache) Delete(ctx context.Context, key string) {
	if err := c.Service.Delete(ctx, key); err != nil && c.log != nil {
		c.log.Warn(ctx, "failed to invalidate story cache entry", "error", err)
	}
}

func (c *storyCache) DeleteByPattern(ctx context.Context, pattern string) {
	if err := c.Service.DeleteByPattern(ctx, pattern); err != nil && c.log != nil {
		c.log.Warn(ctx, "failed to invalidate story cache pattern", "error", err)
	}
}

type storyUserReader interface {
	GetUsersByIDs(ctx context.Context, userIDs []uuid.UUID) ([]users.CoreUser, error)
}

// New creates a new Handlers instance with the required dependencies.
func New(stories *stories.Service, users *users.Service, links *links.Service, attachments *attachments.Service, cacheService *cache.Service, log *logger.Logger) *Handlers {
	return &Handlers{
		stories:     stories,
		users:       users,
		links:       links,
		attachments: attachments,
		storyMedia:  attachments,
		cache:       newStoryCache(cacheService, log),
		log:         log,
	}
}

func (h *Handlers) resolveUserAvatarURL(ctx context.Context, avatar string) string {
	if h.attachments == nil {
		return avatar
	}
	resolved, err := h.attachments.ResolveProfileImageURL(ctx, avatar, 24*time.Hour)
	if err != nil {
		return ""
	}
	return resolved
}

func (h *Handlers) buildAppUserSummary(ctx context.Context, user users.CoreUser) AppUserSummary {
	user.AvatarURL = h.resolveUserAvatarURL(ctx, user.AvatarURL)
	return toAppUserSummary(user)
}

func (h *Handlers) getStoryUsers(ctx context.Context, userIDs []uuid.UUID) (map[uuid.UUID]AppUserSummary, error) {
	usersByID := make(map[uuid.UUID]AppUserSummary, len(userIDs))
	if len(userIDs) == 0 {
		return usersByID, nil
	}
	if h.users == nil {
		return usersByID, nil
	}

	fetchedUsers, err := h.users.GetUsersByIDs(ctx, userIDs)
	if err != nil {
		return nil, err
	}

	for _, user := range fetchedUsers {
		usersByID[user.ID] = h.buildAppUserSummary(ctx, user)
	}

	return usersByID, nil
}

func collectStoryListUserIDs(story stories.CoreStoryList, userIDs map[uuid.UUID]struct{}) {
	if story.Assignee != nil {
		userIDs[*story.Assignee] = struct{}{}
	}
	if story.Reporter != nil {
		userIDs[*story.Reporter] = struct{}{}
	}

	for _, subStory := range story.SubStories {
		collectStoryListUserIDs(subStory, userIDs)
	}
}

func collectStoryUserIDs(story stories.CoreSingleStory, userIDs map[uuid.UUID]struct{}) {
	if story.Assignee != nil {
		userIDs[*story.Assignee] = struct{}{}
	}
	if story.Reporter != nil {
		userIDs[*story.Reporter] = struct{}{}
	}
	for _, collaboratorID := range story.Collaborators {
		userIDs[collaboratorID] = struct{}{}
	}
	for _, watcherID := range story.WatcherIDs {
		userIDs[watcherID] = struct{}{}
	}

	for _, subStory := range story.SubStories {
		collectStoryListUserIDs(subStory, userIDs)
	}

	for _, association := range story.Associations {
		collectStoryListUserIDs(association.Story, userIDs)
	}
}

func mapUserIDs(values map[uuid.UUID]struct{}) []uuid.UUID {
	userIDs := make([]uuid.UUID, 0, len(values))
	for userID := range values {
		userIDs = append(userIDs, userID)
	}
	return userIDs
}

func (h *Handlers) buildStoryUsersByID(ctx context.Context, story stories.CoreSingleStory) (map[uuid.UUID]AppUserSummary, error) {
	userIDs := make(map[uuid.UUID]struct{})
	collectStoryUserIDs(story, userIDs)
	return h.getStoryUsers(ctx, mapUserIDs(userIDs))
}

func (h *Handlers) buildStoriesUsersByID(ctx context.Context, storyList []stories.CoreStoryList) (map[uuid.UUID]AppUserSummary, error) {
	userIDs := make(map[uuid.UUID]struct{})
	for _, story := range storyList {
		collectStoryListUserIDs(story, userIDs)
	}
	return h.getStoryUsers(ctx, mapUserIDs(userIDs))
}

func (h *Handlers) buildStoryGroupsUsersByID(ctx context.Context, groups []stories.CoreStoryGroup) (map[uuid.UUID]AppUserSummary, error) {
	userIDs := make(map[uuid.UUID]struct{})
	for _, group := range groups {
		for _, story := range group.Stories {
			collectStoryListUserIDs(story, userIDs)
		}
	}
	return h.getStoryUsers(ctx, mapUserIDs(userIDs))
}

func (h *Handlers) respondStory(ctx context.Context, w http.ResponseWriter, story stories.CoreSingleStory, statusCode int) error {
	usersByID, err := h.buildStoryUsersByID(ctx, story)
	if err != nil {
		return err
	}

	return web.Respond(ctx, w, toAppStory(story, usersByID), statusCode)
}

func (h *Handlers) respondCreatedStory(ctx context.Context, w http.ResponseWriter, story stories.CoreSingleStory) error {
	usersByID, err := h.buildStoryUsersByID(ctx, story)
	if err != nil {
		if h.log != nil {
			h.log.Error(
				ctx,
				"created story response user enrichment failed",
				"story_id", story.ID,
				"workspace_id", story.Workspace,
				"error_type", fmt.Sprintf("%T", err),
			)
		}
		usersByID = map[uuid.UUID]AppUserSummary{}
	}

	return web.Respond(ctx, w, toAppStory(story, usersByID), http.StatusCreated)
}

func (h *Handlers) respondStories(ctx context.Context, w http.ResponseWriter, storyList []stories.CoreStoryList, statusCode int) error {
	usersByID, err := h.buildStoriesUsersByID(ctx, storyList)
	if err != nil {
		return err
	}

	return web.Respond(ctx, w, toAppStories(storyList, usersByID), statusCode)
}

func collectCommentUserIDs(comment stories.CoreComment, userIDs map[uuid.UUID]struct{}) {
	userIDs[comment.UserID] = struct{}{}

	for _, subComment := range comment.SubComments {
		collectCommentUserIDs(subComment, userIDs)
	}
}

func (h *Handlers) buildCommentsUsersByID(ctx context.Context, commentList []stories.CoreComment) (map[uuid.UUID]AppUserSummary, error) {
	userIDs := make(map[uuid.UUID]struct{})
	for _, comment := range commentList {
		collectCommentUserIDs(comment, userIDs)
	}
	return h.getStoryUsers(ctx, mapUserIDs(userIDs))
}

func (h *Handlers) respondComment(ctx context.Context, w http.ResponseWriter, comment stories.CoreComment, statusCode int) error {
	usersByID, err := h.buildCommentsUsersByID(ctx, []stories.CoreComment{comment})
	if err != nil {
		return err
	}

	return web.Respond(ctx, w, toAppComment(comment, usersByID), statusCode)
}

func (h *Handlers) respondComments(ctx context.Context, w http.ResponseWriter, commentList []stories.CoreComment, response CommentsResponse, statusCode int) error {
	usersByID, err := h.buildCommentsUsersByID(ctx, commentList)
	if err != nil {
		return err
	}

	response.Comments = toAppComments(commentList, usersByID)
	return web.Respond(ctx, w, response, statusCode)
}
