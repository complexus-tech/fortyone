package feedback

import (
	"crypto/rand"
	"fmt"
	"io"
	"regexp"
	"strings"
	"time"

	feedbackdomain "github.com/complexus-tech/projects-api/internal/modules/feedback/domain"
	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/google/uuid"
)

var (
	ErrInvalidInput              = feedbackdomain.ErrInvalidInput
	ErrNotFound                  = feedbackdomain.ErrNotFound
	ErrBoardExists               = feedbackdomain.ErrBoardExists
	ErrAlreadyPlanned            = feedbackdomain.ErrAlreadyPlanned
	ErrTeamMismatch              = feedbackdomain.ErrTeamMismatch
	ErrStoryManaged              = feedbackdomain.ErrStoryManaged
	ErrDuplicateItem             = feedbackdomain.ErrDuplicateItem
	ErrVersionConflict           = feedbackdomain.ErrVersionConflict
	ErrParticipationNotAllowed   = feedbackdomain.ErrParticipationNotAllowed
	ErrAuthenticationRequired    = feedbackdomain.ErrAuthenticationRequired
	ErrContributorSessionInvalid = feedbackdomain.ErrContributorSessionInvalid
	ErrContributorBlocked        = feedbackdomain.ErrContributorBlocked
	ErrVerificationExpired       = feedbackdomain.ErrVerificationExpired
	ErrVerificationConsumed      = feedbackdomain.ErrVerificationConsumed
	ErrVerificationAttempts      = feedbackdomain.ErrVerificationAttempts
	ErrWidgetOriginNotAllowed    = feedbackdomain.ErrWidgetOriginNotAllowed
	ErrWidgetAssertionInvalid    = feedbackdomain.ErrWidgetAssertionInvalid
	ErrWidgetAssertionReplayed   = feedbackdomain.ErrWidgetAssertionReplayed
	ErrFeatureUnavailable        = feedbackdomain.ErrFeatureUnavailable
	ErrMergeConflict             = feedbackdomain.ErrMergeConflict
)

var nonSlugCharacters = regexp.MustCompile(`[^a-z0-9]+`)

const (
	maxPublicFeedbackTitleCharacters           = 200
	maxPublicFeedbackDescriptionCharacters     = 20_000
	maxPublicFeedbackDescriptionHTMLCharacters = 100_000
	maxPublicFeedbackCommentCharacters         = 10_000
	defaultContributorPageSize                 = 20
	maxContributorPageSize                     = 50
	publicRoadmapSlug                          = "roadmap"
	defaultSimilarItemsLimit                   = 3
	maxSimilarItemsLimit                       = 5
	minimumSimilarityTitleCharacters           = 10
	duplicateItemConfidence                    = 0.82
)

type Service struct {
	repo                     Repository
	nextRepo                 NextPhaseRepository
	scopedNextRepo           ScopedNextPhaseStore
	scopedCoreRepo           ScopedCoreStore
	stories                  StoryPlanner
	publisher                EventPublisher
	log                      *logger.Logger
	security                 *contributorSecurity
	websiteURL               string
	tasks                    ContributorTasks
	guestNotificationActorID uuid.UUID
	now                      func() time.Time
	random                   io.Reader
}

type Option func(*Service)

func WithEventPublisher(log *logger.Logger, publisher EventPublisher) Option {
	return func(service *Service) {
		service.log = log
		service.publisher = publisher
	}
}

// WithGuestNotificationActor supplies the existing system actor used as the
// database actor for account notifications initiated by a contributor who has
// no global user row. The payload still carries the contributor's public name.
func WithGuestNotificationActor(actorID uuid.UUID) Option {
	return func(service *Service) {
		service.guestNotificationActorID = actorID
	}
}

func New(repo Repository, stories StoryPlanner, options ...Option) *Service {
	service := &Service{repo: repo, stories: stories, now: time.Now, random: rand.Reader}
	if nextRepo, ok := repo.(NextPhaseRepository); ok {
		service.nextRepo = nextRepo
	}
	if scopedNextRepo, ok := repo.(ScopedNextPhaseStore); ok {
		service.scopedNextRepo = scopedNextRepo
	}
	if scopedCoreRepo, ok := repo.(ScopedCoreStore); ok {
		service.scopedCoreRepo = scopedCoreRepo
	}
	for _, option := range options {
		option(service)
	}
	return service
}

func invalidInput(message string) error {
	return fmt.Errorf("%w: %s", ErrInvalidInput, message)
}

func invalidInputf(format string, args ...any) error {
	return fmt.Errorf("%w: %s", ErrInvalidInput, fmt.Sprintf(format, args...))
}

func normalizeSlug(value string) string {
	slug := strings.ToLower(strings.TrimSpace(value))
	slug = nonSlugCharacters.ReplaceAllString(slug, "-")
	return strings.Trim(slug, "-")
}

func isValidStatus(status string) bool {
	switch status {
	case StatusPending, StatusReviewing, StatusPlanned, StatusInProgress, StatusCompleted, StatusClosed:
		return true
	default:
		return false
	}
}

func isValidSubmissionSource(source string) bool {
	switch source {
	case SubmissionSourceInternal, SubmissionSourcePortal, SubmissionSourceWidget, SubmissionSourceIntegration:
		return true
	default:
		return false
	}
}

func isValidReviewerEmailFrequency(frequency string) bool {
	switch frequency {
	case EmailFrequencyOff, EmailFrequencyDaily, EmailFrequencyWeekly:
		return true
	default:
		return false
	}
}

func isValidParticipationMode(mode string) bool {
	return mode == ParticipationModeAccountRequired || mode == ParticipationModeVerifiedGuest || mode == ParticipationModeAnonymousAllowed
}

func isValidGuestIdentityPolicy(policy string) bool {
	switch policy {
	case GuestIdentityPolicyShowIdentity, GuestIdentityPolicyAllowPublicMasking, GuestIdentityPolicyAlwaysMaskGuests:
		return true
	default:
		return false
	}
}

func pointer[T any](value T) *T {
	return &value
}

func coreItemIDs(items []CoreItem) []uuid.UUID {
	ids := make([]uuid.UUID, 0, len(items))
	for _, item := range items {
		ids = append(ids, item.ID)
	}
	return ids
}
