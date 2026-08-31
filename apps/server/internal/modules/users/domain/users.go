package usersdomain

import (
	"errors"
	"sort"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"
)

var (
	ErrNotFound              = errors.New("we couldn't find your account")
	ErrEmailTaken            = errors.New("the email address is already registered")
	ErrTokenExpired          = errors.New("the sign-in link has expired - please request a new one")
	ErrTokenUsed             = errors.New("the sign-in link has already been used")
	ErrTooManyAttempts       = errors.New("too many sign-in attempts - please wait a few minutes and try again")
	ErrInvalidToken          = errors.New("the sign-in link is invalid - please request a new one")
	ErrUserNotFound          = errors.New("user not found")
	ErrInvalidCredentials    = errors.New("invalid credentials")
	ErrEmailAlreadyExists    = errors.New("email already exists")
	ErrUsernameAlreadyExists = errors.New("username already exists")
	ErrTokenNotFound         = errors.New("token not found")
	ErrTokenCollision        = errors.New("verification token collision")
	ErrVerificationDisabled  = errors.New("verification token service is not configured")
	ErrWorkspaceNotFound     = errors.New("workspace not found")
	ErrMemoryNotFound        = errors.New("memory not found")
	ErrIdentityStoreMissing  = errors.New("external identity storage is not configured")
	ErrInvalidOnboardingTour = errors.New("invalid onboarding tour progress")
)

type VerificationToken struct {
	ID           uuid.UUID  `json:"id"`
	Token        string     `json:"token,omitempty"`
	Email        string     `json:"email"`
	UserID       *uuid.UUID `json:"userId"`
	ExpiresAt    time.Time  `json:"expiresAt"`
	UsedAt       *time.Time `json:"usedAt"`
	TokenType    string     `json:"tokenType"`
	TokenKeyID   string     `json:"-"`
	TokenVersion int16      `json:"-"`
	CreatedAt    time.Time  `json:"createdAt"`
	UpdatedAt    time.Time  `json:"updatedAt"`
}

// NewVerificationToken is the digest-only persistence contract for a newly
// issued code. Token plaintext must never cross the repository boundary.
type NewVerificationToken struct {
	Email          string
	TokenType      string
	TokenDigest    []byte
	TokenKeyID     string
	TokenVersion   int16
	ExpiresAt      time.Time
	IssuedAt       time.Time
	RateLimitSince time.Time
	MaximumIssues  int
}

// ConsumeVerificationToken contains every approved purpose-bound digest for
// one verification attempt. LegacyToken exists only for the bounded schema
// expand/contract window.
type ConsumeVerificationToken struct {
	Email         string
	TokenTypes    []string
	TokenDigests  [][]byte
	TokenKeyIDs   []string
	TokenVersions []int16
	LegacyToken   string
	ConsumedAt    time.Time
}

type User struct {
	ID                            uuid.UUID
	Username                      string
	Email                         string
	FullName                      string
	AvatarURL                     string
	IsActive                      bool
	IsSystem                      bool
	IsInternal                    bool
	HasSeenWalkthrough            bool
	Timezone                      string
	WorkingDays                   []int
	WorkingStartMinute            *int
	WorkingEndMinute              *int
	LastLoginAt                   time.Time
	LastUsedWorkspaceID           *uuid.UUID
	GitHubUsername                *string
	CreatedAt                     time.Time
	UpdatedAt                     time.Time
	Token                         *string
	Role                          *string
	TeamAIRoleTitle               string
	TeamAIRoleDescription         string
	InferredTeamAIRoleTitle       string
	InferredTeamAIRoleDescription string
	InferredTeamAIRoleStoryCount  int
	InferredTeamAIRoleConfidence  float32
	InferredTeamAIRoleGeneratedAt *time.Time
	LastStoryActivityAt           *time.Time
}

// Login reactivation policies are persisted account security decisions rather
// than presentation state. Only VerifiedSignIn permits an inactive account to
// become active as a consequence of a successfully verified authentication.
type LoginReactivationPolicy string

const (
	LoginReactivationVerifiedSignIn LoginReactivationPolicy = "verified_sign_in"
	LoginReactivationAdminOnly      LoginReactivationPolicy = "admin_only"
	LoginReactivationLegacyReview   LoginReactivationPolicy = "legacy_admin_review"
)

type VerifiedSignInReactivation struct {
	UserID     uuid.UUID
	SignedInAt time.Time
}

type ListUsersFilter struct {
	TeamID *uuid.UUID
	Search string
	Limit  int
	Offset int
}

type UpdateUser struct {
	Username           *string
	FullName           *string
	AvatarURL          *string
	HasSeenWalkthrough *bool
	Timezone           *string
	WorkSchedule       *WorkScheduleOverride
}

type WorkScheduleOverride struct {
	WorkingDays        []int
	WorkingStartMinute *int
	WorkingEndMinute   *int
}

type NewUser struct {
	Email     string
	FullName  string
	AvatarURL string
	Timezone  string
}

type ExternalIdentityInput struct {
	Provider  string
	Issuer    string
	Subject   string
	Email     string
	FullName  string
	AvatarURL string
	Timezone  string
}

type ExternalIdentityResult struct {
	User    User
	Created bool
}

type AutomationPreferences struct {
	UserID                     uuid.UUID
	WorkspaceID                uuid.UUID
	AutoAssignSelf             bool
	AutoScheduling             bool
	AssignSelfOnBranchCopy     bool
	MoveStoryToStartedOnBranch bool
	OpenStoryInDialog          bool
	CreatedAt                  time.Time
	UpdatedAt                  time.Time
}

type UpdateAutomationPreferences struct {
	AutoAssignSelf             *bool
	AutoScheduling             *bool
	AssignSelfOnBranchCopy     *bool
	MoveStoryToStartedOnBranch *bool
	OpenStoryInDialog          *bool
}

const (
	MaximumOnboardingTourKeyRunes        = 64
	MaximumOnboardingTourVersionRunes    = 64
	MaximumOnboardingTourProgressIDs     = 32
	MaximumOnboardingTourProgressIDRunes = 96
)

type OnboardingTourStatus string

const (
	OnboardingTourStatusActive    OnboardingTourStatus = "active"
	OnboardingTourStatusCompleted OnboardingTourStatus = "completed"
	OnboardingTourStatusSkipped   OnboardingTourStatus = "skipped"
)

type OnboardingTourScope struct {
	TourKey     string
	TourVersion string
}

type OnboardingTourProgress struct {
	UserID             uuid.UUID
	WorkspaceID        uuid.UUID
	TourKey            string
	TourVersion        string
	CompletedStepIDs   []string
	CompletedActionIDs []string
	Status             OnboardingTourStatus
	CreatedAt          time.Time
	UpdatedAt          time.Time
}

type UpdateOnboardingTourProgress struct {
	OnboardingTourScope
	CompletedStepIDs   []string
	CompletedActionIDs []string
	Status             *OnboardingTourStatus
}

func (scope *OnboardingTourScope) NormalizeAndValidate() error {
	if scope == nil {
		return ErrInvalidOnboardingTour
	}

	tourKey, err := normalizeOnboardingTourText(scope.TourKey, MaximumOnboardingTourKeyRunes)
	if err != nil {
		return err
	}
	tourVersion, err := normalizeOnboardingTourText(scope.TourVersion, MaximumOnboardingTourVersionRunes)
	if err != nil {
		return err
	}

	scope.TourKey = tourKey
	scope.TourVersion = tourVersion
	return nil
}

func (update *UpdateOnboardingTourProgress) NormalizeAndValidate() error {
	if update == nil {
		return ErrInvalidOnboardingTour
	}
	if err := update.OnboardingTourScope.NormalizeAndValidate(); err != nil {
		return err
	}

	completedStepIDs, err := normalizeOnboardingTourProgressIDs(update.CompletedStepIDs)
	if err != nil {
		return err
	}
	completedActionIDs, err := normalizeOnboardingTourProgressIDs(update.CompletedActionIDs)
	if err != nil {
		return err
	}

	update.CompletedStepIDs = completedStepIDs
	update.CompletedActionIDs = completedActionIDs

	if update.Status == nil {
		return nil
	}
	status := OnboardingTourStatus(strings.TrimSpace(string(*update.Status)))
	if !status.valid() {
		return ErrInvalidOnboardingTour
	}
	update.Status = &status
	return nil
}

func (status OnboardingTourStatus) valid() bool {
	return status == OnboardingTourStatusActive ||
		status == OnboardingTourStatusCompleted ||
		status == OnboardingTourStatusSkipped
}

func normalizeOnboardingTourText(value string, maximumRunes int) (string, error) {
	value = strings.TrimSpace(value)
	if value == "" || !utf8.ValidString(value) || strings.ContainsRune(value, '\x00') || utf8.RuneCountInString(value) > maximumRunes {
		return "", ErrInvalidOnboardingTour
	}
	return value, nil
}

func normalizeOnboardingTourProgressIDs(values []string) ([]string, error) {
	if len(values) > MaximumOnboardingTourProgressIDs {
		return nil, ErrInvalidOnboardingTour
	}

	seen := make(map[string]struct{}, len(values))
	normalized := make([]string, 0, len(values))
	for _, value := range values {
		identifier, err := normalizeOnboardingTourText(value, MaximumOnboardingTourProgressIDRunes)
		if err != nil {
			return nil, err
		}
		if _, exists := seen[identifier]; exists {
			continue
		}
		seen[identifier] = struct{}{}
		normalized = append(normalized, identifier)
	}
	sort.Strings(normalized)
	return normalized, nil
}

type UserMemoryItem struct {
	ID          uuid.UUID `json:"id"`
	UserID      uuid.UUID `json:"userId"`
	WorkspaceID uuid.UUID `json:"workspaceId"`
	Content     string    `json:"content"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

type NewUserMemoryItem struct {
	UserID      uuid.UUID
	WorkspaceID uuid.UUID
	Content     string
}

type UserMemoryScope struct {
	UserID      uuid.UUID
	WorkspaceID uuid.UUID
}

type UpdateUserMemoryItem struct {
	Content *string
}
