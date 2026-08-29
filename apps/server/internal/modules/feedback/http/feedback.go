package feedbackhttp

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"time"

	feedback "github.com/complexus-tech/projects-api/internal/modules/feedback/service"
	teams "github.com/complexus-tech/projects-api/internal/modules/teams/service"
	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/complexus-tech/projects-api/pkg/web"
	"github.com/google/uuid"
)

const (
	avatarAccessURLExpiry          = 24 * time.Hour
	publicFeedbackItemBodyLimit    = 128 << 10
	publicFeedbackCommentBodyLimit = 64 << 10
	publicFeedbackVoteBodyLimit    = 4 << 10
	defaultTeamFeedbackPageSize    = 25
)

func decodePublicRequest(w http.ResponseWriter, r *http.Request, input any, bodyLimit int64) (int, error) {
	r.Body = http.MaxBytesReader(w, r.Body, bodyLimit)
	if err := web.Decode(r, input); err != nil {
		if errors.Is(err, web.ErrRequestBodyTooLarge) {
			return http.StatusRequestEntityTooLarge, err
		}
		return http.StatusBadRequest, err
	}
	return 0, nil
}

func validatePublicItemBotTrap(input AppCreatePublicItem) error {
	if strings.TrimSpace(input.Website) != "" {
		return errors.New("invalid feedback submission")
	}
	return nil
}

type profileImageResolver interface {
	ResolveProfileImageURL(ctx context.Context, avatar string, expiry time.Duration) (string, error)
}

type teamAccessService interface {
	GetByID(ctx context.Context, teamID, workspaceID, userID uuid.UUID) (teams.CoreTeam, error)
}

type Handlers struct {
	feedback      *feedback.Service
	teams         teamAccessService
	profileImages profileImageResolver
	log           *logger.Logger
}

func New(service *feedback.Service, teamAccess teamAccessService, profileImages profileImageResolver, log *logger.Logger) *Handlers {
	return &Handlers{feedback: service, teams: teamAccess, profileImages: profileImages, log: log}
}

func (h *Handlers) authorizeTeam(ctx context.Context, workspaceID, teamID, userID uuid.UUID) error {
	if h.teams == nil {
		return errors.New("team access service is required")
	}
	if _, err := h.teams.GetByID(ctx, teamID, workspaceID, userID); err != nil {
		if h.log != nil {
			h.log.Warn(ctx, "feedback team access denied", "team_id", teamID, "user_id", userID, "error", err)
		}
		if errors.Is(err, teams.ErrTeamNotFound) {
			return feedback.ErrNotFound
		}
		return err
	}
	return nil
}

func (h *Handlers) authorizeItemTeam(ctx context.Context, workspaceID, itemID, userID uuid.UUID) error {
	item, err := h.feedback.GetItem(ctx, workspaceID, itemID)
	if err != nil {
		return err
	}
	return h.authorizeTeam(ctx, workspaceID, item.Board.TeamID, userID)
}

func (h *Handlers) resolveAuthorAvatar(
	ctx context.Context,
	avatar *string,
	resolvedByAvatar map[string]*string,
) *string {
	if avatar == nil {
		return nil
	}

	avatarKey := strings.TrimSpace(*avatar)
	if avatarKey == "" {
		return nil
	}
	if resolved, ok := resolvedByAvatar[avatarKey]; ok {
		return resolved
	}
	if h.profileImages == nil {
		resolvedByAvatar[avatarKey] = nil
		return nil
	}

	resolved, err := h.profileImages.ResolveProfileImageURL(ctx, avatarKey, avatarAccessURLExpiry)
	if err != nil {
		if h.log != nil {
			h.log.Warn(ctx, "failed to resolve feedback author avatar", "error", err)
		}
		resolvedByAvatar[avatarKey] = nil
		return nil
	}
	if strings.TrimSpace(resolved) == "" {
		resolvedByAvatar[avatarKey] = nil
		return nil
	}

	resolvedByAvatar[avatarKey] = &resolved
	return &resolved
}

func (h *Handlers) resolvePortalAvatars(ctx context.Context, portal *feedback.CorePortalSnapshot) {
	resolvedByAvatar := make(map[string]*string)
	itemIDs := make(map[uuid.UUID]struct{}, len(portal.Items))

	for i := range portal.Items {
		itemIDs[portal.Items[i].ID] = struct{}{}
		portal.Items[i].AuthorAvatar = h.resolveAuthorAvatar(ctx, portal.Items[i].AuthorAvatar, resolvedByAvatar)
	}
	for i := range portal.Comments {
		if _, visible := itemIDs[portal.Comments[i].ItemID]; !visible {
			continue
		}
		portal.Comments[i].AuthorAvatar = h.resolveAuthorAvatar(ctx, portal.Comments[i].AuthorAvatar, resolvedByAvatar)
	}
}

func (h *Handlers) resolveSimilarItemAvatars(ctx context.Context, items []feedback.CoreSimilarItem) {
	resolvedByAvatar := make(map[string]*string)
	for i := range items {
		items[i].AuthorAvatar = h.resolveAuthorAvatar(ctx, items[i].AuthorAvatar, resolvedByAvatar)
	}
}

func httpStatus(err error) int {
	switch {
	case errors.Is(err, feedback.ErrNotFound):
		return http.StatusNotFound
	case errors.Is(err, feedback.ErrAlreadyPlanned):
		return http.StatusConflict
	case errors.Is(err, feedback.ErrBoardExists):
		return http.StatusConflict
	case errors.Is(err, feedback.ErrStoryManaged):
		return http.StatusConflict
	case errors.Is(err, feedback.ErrDuplicateItem):
		return http.StatusConflict
	case errors.Is(err, feedback.ErrParticipationNotAllowed):
		return http.StatusForbidden
	case errors.Is(err, feedback.ErrContributorBlocked), errors.Is(err, feedback.ErrWidgetOriginNotAllowed):
		return http.StatusForbidden
	case errors.Is(err, feedback.ErrAuthenticationRequired), errors.Is(err, feedback.ErrContributorSessionInvalid), errors.Is(err, feedback.ErrWidgetAssertionInvalid):
		return http.StatusUnauthorized
	case errors.Is(err, feedback.ErrVerificationExpired), errors.Is(err, feedback.ErrVerificationConsumed):
		return http.StatusGone
	case errors.Is(err, feedback.ErrVerificationAttempts):
		return http.StatusTooManyRequests
	case errors.Is(err, feedback.ErrWidgetAssertionReplayed):
		return http.StatusConflict
	case errors.Is(err, feedback.ErrMergeConflict):
		return http.StatusConflict
	case errors.Is(err, feedback.ErrFeatureUnavailable):
		return http.StatusServiceUnavailable
	case errors.Is(err, feedback.ErrTeamMismatch):
		return http.StatusBadRequest
	case errors.Is(err, feedback.ErrInvalidInput):
		return http.StatusBadRequest
	default:
		return http.StatusInternalServerError
	}
}
