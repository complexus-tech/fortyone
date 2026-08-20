package calendar

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/url"
	"strings"
	"time"

	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/google/uuid"
	"google.golang.org/api/googleapi"
)

var (
	ErrCalendarNotConfigured           = errors.New("calendar integration is not configured")
	ErrInvalidCalendarState            = errors.New("invalid calendar setup state")
	ErrCalendarNotFound                = errors.New("calendar connection not found")
	ErrCalendarAccessDenied            = errors.New("calendar access denied")
	ErrCalendarCredentialsIncomplete   = errors.New("calendar credentials are incomplete")
	ErrCalendarEventNotFound           = errors.New("calendar event not found")
	ErrCalendarSyncSuperseded          = errors.New("calendar sync was superseded by newer credentials")
	ErrInvalidScheduleRange            = errors.New("calendar schedule range is invalid")
	ErrInvalidScheduleBlock            = errors.New("calendar schedule block is invalid")
	ErrCalendarScheduleConflict        = errors.New("calendar time conflicts with an existing meeting or schedule block")
	ErrCalendarScheduleStalePlan       = errors.New("calendar schedule plan is stale")
	ErrCalendarScheduleBlockNotFound   = errors.New("calendar schedule block not found")
	ErrManagedScheduleBlock            = errors.New("Maya-managed schedule blocks can only be changed by automatic scheduling")
	ErrCalendarCleanupPending          = errors.New("calendar cleanup is still in progress")
	ErrCalendarAccountChangePending    = errors.New("disconnect the existing calendar before connecting a different Google account")
	ErrCalendarFullSyncRequired        = errors.New("calendar full sync is required")
	ErrInvalidCalendarNotification     = errors.New("invalid calendar notification")
	ErrCalendarWebhookNotConfigured    = errors.New("calendar webhook is not configured")
	ErrCalendarReauthorizationRequired = errors.New("calendar write access requires reauthorization")
)

const (
	connectStateTTL                    = 10 * time.Minute
	defaultSyncLookback                = -7 * 24 * time.Hour
	defaultSyncLookahead               = 90 * 24 * time.Hour
	googleWatchTTL                     = 7 * 24 * time.Hour
	googleWatchRenewalWindow           = 24 * time.Hour
	maximumScheduleEventOutboxAttempts = 8
)

type CalendarTaskQueue interface {
	EnqueueCalendarSync(ctx context.Context, connectionID uuid.UUID) error
}

type CalendarScheduleTaskQueue interface {
	EnqueueCalendarScheduleReconcile(ctx context.Context, userID uuid.UUID) error
}

type CalendarUpdatePublisher interface {
	PublishCalendarUpdated(ctx context.Context, workspaceID, userID, connectionID uuid.UUID, syncedAt time.Time) error
}

type Repository interface {
	ListConnections(ctx context.Context, workspaceID uuid.UUID, userID *uuid.UUID) ([]CoreConnection, error)
	GetOwnedConnection(ctx context.Context, workspaceID, userID, connectionID uuid.UUID) (CoreConnection, error)
	GetActiveConnection(ctx context.Context, workspaceID, userID uuid.UUID, provider Provider) (CoreConnection, error)
	GetConnection(ctx context.Context, connectionID uuid.UUID) (CoreConnection, error)
	GetScheduleEventDispatchConnection(ctx context.Context, userID uuid.UUID) (CoreConnection, bool, error)
	ListConnectionsNeedingWatch(ctx context.Context, renewBefore time.Time) ([]CoreConnection, error)
	WorkspaceMemberExists(ctx context.Context, workspaceID, userID uuid.UUID) (bool, error)
	UpsertConnection(ctx context.Context, input CoreConnectionUpsert) (CoreConnection, error)
	BeginConnectionSync(ctx context.Context, connection CoreConnection) (CoreConnection, error)
	RevokeConnection(ctx context.Context, workspaceID, userID, connectionID uuid.UUID) error
	ReplaceCalendarSnapshot(ctx context.Context, connection CoreConnection, snapshot CalendarSyncSnapshot) error
	ApplyCalendarChanges(ctx context.Context, connection CoreConnection, delta CalendarSyncDelta) error
	SetNotificationChannel(ctx context.Context, connection CoreConnection, channel CalendarWatchChannel) error
	ClearNotificationChannel(ctx context.Context, connectionID uuid.UUID) error
	ListCalendarEvents(ctx context.Context, workspaceID, userID uuid.UUID, startAt, endAt time.Time) ([]CoreCalendarEventSummary, error)
	GetCalendarEvent(ctx context.Context, workspaceID, userID, eventID uuid.UUID) (CoreCalendarEvent, error)
	ListBusyWindows(ctx context.Context, workspaceID, userID uuid.UUID, startAt, endAt time.Time) ([]CoreBusyWindow, error)
	ListScheduleBlocks(ctx context.Context, workspaceID, userID uuid.UUID, startAt, endAt time.Time) ([]CoreScheduleBlock, error)
	GetScheduleBlock(ctx context.Context, workspaceID, userID, blockID uuid.UUID) (CoreScheduleBlock, error)
	ScheduleStoryExists(ctx context.Context, workspaceID, userID, storyID uuid.UUID) (bool, error)
	CreateScheduleBlock(ctx context.Context, input CoreScheduleBlockInput) (CoreScheduleBlock, error)
	UpdateScheduleBlock(ctx context.Context, input CoreScheduleBlockInput) (CoreScheduleBlock, error)
	DeleteScheduleBlock(ctx context.Context, workspaceID, userID, blockID uuid.UUID) error
	MarkConnectionSynced(ctx context.Context, workspaceID, connectionID, credentialGeneration uuid.UUID, syncedAt time.Time) error
	MarkConnectionSyncFailed(ctx context.Context, workspaceID, connectionID, credentialGeneration uuid.UUID, message string) error
}

type ScheduleReconciliationRepository interface {
	ListSchedulingBlocksForUser(ctx context.Context, workspaceID, userID uuid.UUID, startAt, endAt time.Time) ([]CoreScheduleBlock, error)
	ListMayaScheduleBlocksForStory(ctx context.Context, workspaceID, userID, storyID uuid.UUID) ([]CoreScheduleBlock, error)
	MayaScheduleOwnershipExists(ctx context.Context, workspaceID, userID, storyID uuid.UUID) (bool, error)
	ReconcileMayaScheduleBlocks(ctx context.Context, input MayaScheduleReconcileInput) (CoreScheduleReconcileResult, error)
	ListReadyScheduleEventOutboxUsers(ctx context.Context, limit int) ([]uuid.UUID, error)
	WithScheduleEventDispatchLock(ctx context.Context, userID uuid.UUID, dispatch func(ScheduleEventOutboxStore) error) error
}

type ScheduleIssueRepository interface {
	ListScheduleIssues(ctx context.Context, workspaceID, userID uuid.UUID) ([]CoreScheduleIssue, error)
}

type ScheduleEventOutboxStore interface {
	ListPendingScheduleEventOutbox(ctx context.Context, userID uuid.UUID, limit int) ([]CoreScheduleEventOutbox, error)
	ScheduleEventUpsertIsCurrent(ctx context.Context, item CoreScheduleEventOutbox, event ExternalScheduleEventInput) (bool, error)
	MarkScheduleEventOutboxProcessed(ctx context.Context, item CoreScheduleEventOutbox, syncHash string) error
	MarkScheduleEventOutboxFailed(ctx context.Context, item CoreScheduleEventOutbox, message string, permanent bool) error
	ReleaseScheduleEventOutbox(ctx context.Context, outboxIDs []uuid.UUID) error
	DeleteCleanupPendingConnectionIfDrained(ctx context.Context, userID uuid.UUID) error
}

type Config struct {
	SecretKey      string
	WebsiteURL     string
	WebhookURL     string
	RequireWebhook bool
	Providers      map[Provider]CalendarProvider
	Tasks          CalendarTaskQueue
	Updates        CalendarUpdatePublisher
}

type Service struct {
	log       *logger.Logger
	repo      Repository
	cfg       Config
	now       func() time.Time
	randBytes func([]byte) (int, error)
}

func New(log *logger.Logger, repo Repository, cfg Config) *Service {
	return &Service{
		log:       log,
		repo:      repo,
		cfg:       cfg,
		now:       time.Now,
		randBytes: rand.Read,
	}
}

func (s *Service) ListConnections(ctx context.Context, workspaceID uuid.UUID, userID *uuid.UUID) ([]CoreConnection, error) {
	if s.repo == nil {
		return nil, ErrCalendarNotConfigured
	}
	return s.repo.ListConnections(ctx, workspaceID, userID)
}

func (s *Service) CreateConnectSession(ctx context.Context, workspaceID, userID uuid.UUID, workspaceSlug string) (CoreConnectSession, error) {
	provider, err := s.provider(ProviderGoogle)
	if err != nil {
		return CoreConnectSession{}, err
	}
	state, err := s.signState(stateClaims{
		WorkspaceID:   workspaceID,
		UserID:        userID,
		WorkspaceSlug: strings.TrimSpace(workspaceSlug),
		Provider:      ProviderGoogle,
		ExpiresAt:     s.now().Add(connectStateTTL).Unix(),
	})
	if err != nil {
		return CoreConnectSession{}, err
	}
	authURL, err := provider.AuthCodeURL(state)
	if err != nil {
		return CoreConnectSession{}, err
	}
	return CoreConnectSession{AuthURL: authURL}, nil
}

func (s *Service) CompleteConnect(
	ctx context.Context,
	callbackUserID uuid.UUID,
	code, state string,
) (CoreConnection, string, error) {
	if s.repo == nil {
		return CoreConnection{}, "", ErrCalendarNotConfigured
	}
	claims, err := s.verifyActiveState(state)
	if err != nil {
		return CoreConnection{}, "", err
	}
	if callbackUserID == uuid.Nil || claims.UserID != callbackUserID {
		return CoreConnection{}, "", ErrCalendarAccessDenied
	}
	isMember, err := s.repo.WorkspaceMemberExists(ctx, claims.WorkspaceID, claims.UserID)
	if err != nil {
		return CoreConnection{}, "", err
	}
	if !isMember {
		return CoreConnection{}, "", ErrCalendarAccessDenied
	}
	provider, err := s.provider(claims.Provider)
	if err != nil {
		return CoreConnection{}, "", err
	}
	token, err := provider.ExchangeCode(ctx, strings.TrimSpace(code))
	if err != nil {
		return CoreConnection{}, "", err
	}
	token, err = s.withRetainedRefreshToken(ctx, claims, token)
	if err != nil {
		return CoreConnection{}, "", err
	}
	payload, err := s.encryptTokenPayload(token)
	if err != nil {
		return CoreConnection{}, "", err
	}
	connection, err := s.repo.UpsertConnection(ctx, CoreConnectionUpsert{
		WorkspaceID:       claims.WorkspaceID,
		UserID:            claims.UserID,
		Provider:          claims.Provider,
		ProviderAccountID: strings.TrimSpace(token.ProviderAccountID),
		ConnectedEmail:    strings.TrimSpace(token.ConnectedEmail),
		Timezone:          fallbackTimezone(token.Timezone),
		TokenPayload:      payload,
		Scopes:            token.Scopes,
	})
	if err != nil {
		return CoreConnection{}, "", err
	}
	if err := s.syncConnection(ctx, connection); err != nil {
		return CoreConnection{}, "", fmt.Errorf("sync calendar connection after connect: %w", err)
	}
	current, err := s.repo.GetConnection(ctx, connection.ID)
	if err != nil {
		return CoreConnection{}, "", fmt.Errorf("load calendar connection after initial sync: %w", err)
	}
	if err := s.replaceNotificationChannel(ctx, current); err != nil {
		return CoreConnection{}, "", fmt.Errorf("start calendar notification channel: %w", err)
	}
	if scheduleTasks, ok := s.cfg.Tasks.(CalendarScheduleTaskQueue); ok {
		if err := scheduleTasks.EnqueueCalendarScheduleReconcile(ctx, current.UserID); err != nil {
			return CoreConnection{}, "", fmt.Errorf("enqueue calendar schedule reconciliation after connect: %w", err)
		}
	}
	return connection, s.workspaceCalendarURL(claims.WorkspaceSlug, "connected=1"), nil
}

func (s *Service) withRetainedRefreshToken(
	ctx context.Context,
	claims stateClaims,
	token ProviderToken,
) (ProviderToken, error) {
	if strings.TrimSpace(token.AccessToken) == "" ||
		strings.TrimSpace(token.ProviderAccountID) == "" ||
		strings.TrimSpace(token.ConnectedEmail) == "" {
		return ProviderToken{}, ErrCalendarCredentialsIncomplete
	}
	if strings.TrimSpace(token.RefreshToken) != "" {
		return token, nil
	}

	existing, err := s.repo.GetActiveConnection(ctx, claims.WorkspaceID, claims.UserID, claims.Provider)
	if err != nil {
		if errors.Is(err, ErrCalendarNotFound) {
			return ProviderToken{}, ErrCalendarCredentialsIncomplete
		}
		return ProviderToken{}, err
	}
	if strings.TrimSpace(existing.ProviderAccountID) == "" ||
		existing.ProviderAccountID != token.ProviderAccountID {
		return ProviderToken{}, ErrCalendarCredentialsIncomplete
	}
	previousToken, err := s.decryptTokenPayload(existing.TokenPayload)
	if err != nil || strings.TrimSpace(previousToken.RefreshToken) == "" {
		return ProviderToken{}, ErrCalendarCredentialsIncomplete
	}
	token.RefreshToken = previousToken.RefreshToken
	return token, nil
}

// CalendarCallbackErrorURL turns a valid signed OAuth state into a safe
// account-settings redirect without reflecting provider error text.
func (s *Service) CalendarCallbackErrorURL(state string, callbackUserID uuid.UUID, code string) (string, error) {
	claims, err := s.verifyActiveState(state)
	if err != nil {
		return "", err
	}
	if callbackUserID == uuid.Nil || claims.UserID != callbackUserID {
		return "", ErrCalendarAccessDenied
	}
	switch code {
	case "access_denied", "connection_failed":
	default:
		code = "connection_failed"
	}
	return s.workspaceCalendarURL(
		claims.WorkspaceSlug,
		"calendar_error="+url.QueryEscape(code),
	), nil
}

func (s *Service) RevokeConnection(ctx context.Context, workspaceID, userID, connectionID uuid.UUID) error {
	if s.repo == nil {
		return ErrCalendarNotConfigured
	}
	connection, err := s.repo.GetOwnedConnection(ctx, workspaceID, userID, connectionID)
	if err != nil {
		return err
	}
	s.stopNotificationChannel(ctx, connection)
	if err := s.repo.RevokeConnection(ctx, workspaceID, userID, connectionID); err != nil {
		return err
	}
	// Provider deletion is backed by the durable outbox. Try immediately for a
	// responsive disconnect, while the periodic dispatcher remains the retry
	// contract if Google is temporarily unavailable.
	_ = s.DispatchScheduleEventOutbox(ctx, userID)
	return nil
}

func (s *Service) ProcessGoogleNotification(ctx context.Context, channelID, resourceID, state, token string) error {
	connectionID, tokenChannelID, err := s.verifyNotificationToken(token)
	if err != nil || strings.TrimSpace(channelID) != tokenChannelID {
		return ErrInvalidCalendarNotification
	}
	if state == "sync" {
		return nil
	}
	if state != "exists" && state != "not_exists" {
		return ErrInvalidCalendarNotification
	}
	connection, err := s.repo.GetConnection(ctx, connectionID)
	if err != nil {
		if errors.Is(err, ErrCalendarNotFound) {
			return nil
		}
		return err
	}
	if connection.NotificationChannelID != channelID || connection.NotificationResourceID != strings.TrimSpace(resourceID) {
		return nil
	}
	if s.cfg.Tasks == nil {
		return ErrCalendarNotConfigured
	}
	return s.cfg.Tasks.EnqueueCalendarSync(ctx, connectionID)
}

func (s *Service) SyncConnectionFromNotification(ctx context.Context, connectionID uuid.UUID) error {
	connection, err := s.repo.GetConnection(ctx, connectionID)
	if err != nil {
		if errors.Is(err, ErrCalendarNotFound) {
			return nil
		}
		return err
	}
	if strings.TrimSpace(connection.SyncToken) == "" || !connection.CanReadEventDetails() {
		if err := s.syncConnection(ctx, connection); err != nil {
			return err
		}
		return s.enqueueScheduleReconciliation(ctx, connection.UserID)
	}
	connection, err = s.repo.BeginConnectionSync(ctx, connection)
	if err != nil {
		return err
	}
	provider, err := s.provider(connection.Provider)
	if err != nil {
		return s.failConnectionSync(ctx, connection, err)
	}
	token, err := s.decryptTokenPayload(connection.TokenPayload)
	if err != nil {
		return s.failConnectionSync(ctx, connection, err)
	}
	delta, err := provider.SyncCalendarChanges(ctx, token, connection.SyncToken)
	if errors.Is(err, ErrCalendarFullSyncRequired) {
		current, getErr := s.repo.GetConnection(ctx, connection.ID)
		if getErr != nil {
			return getErr
		}
		if err := s.syncConnection(ctx, current); err != nil {
			return err
		}
		return s.enqueueScheduleReconciliation(ctx, current.UserID)
	}
	if err != nil {
		return s.failConnectionSync(ctx, connection, err)
	}
	s.normalizeDelta(connection, &delta)
	if err := s.repo.ApplyCalendarChanges(ctx, connection, delta); err != nil {
		return s.failConnectionSync(ctx, connection, err)
	}
	if err := s.completeConnectionSync(ctx, connection); err != nil {
		return err
	}
	// Always enqueue after a successful incremental sync, including an empty
	// delta. If the previous attempt committed Google's next sync token but the
	// enqueue failed, Asynq retries with an empty delta; this trailing enqueue is
	// what makes that handoff retry-safe.
	return s.enqueueScheduleReconciliation(ctx, connection.UserID)
}

func (s *Service) enqueueScheduleReconciliation(ctx context.Context, userID uuid.UUID) error {
	scheduleTasks, ok := s.cfg.Tasks.(CalendarScheduleTaskQueue)
	if !ok {
		return nil
	}
	return scheduleTasks.EnqueueCalendarScheduleReconcile(ctx, userID)
}

func (s *Service) RenewExpiringNotificationChannels(ctx context.Context) (int, error) {
	if strings.TrimSpace(s.cfg.WebhookURL) == "" {
		return 0, nil
	}
	connections, err := s.repo.ListConnectionsNeedingWatch(ctx, s.now().Add(googleWatchRenewalWindow))
	if err != nil {
		return 0, err
	}
	renewed := 0
	var renewalErr error
	for _, connection := range connections {
		if err := s.replaceNotificationChannel(ctx, connection); err != nil {
			renewalErr = errors.Join(renewalErr, err)
			continue
		}
		renewed++
	}
	return renewed, renewalErr
}

func (s *Service) SyncConnection(ctx context.Context, workspaceID, userID, connectionID uuid.UUID) error {
	if s.repo == nil {
		return ErrCalendarNotConfigured
	}
	connection, err := s.repo.GetOwnedConnection(ctx, workspaceID, userID, connectionID)
	if err != nil {
		return err
	}
	if err := s.syncConnection(ctx, connection); err != nil {
		return err
	}
	return s.enqueueScheduleReconciliation(ctx, connection.UserID)
}

func (s *Service) SyncActiveGoogleConnection(ctx context.Context, workspaceID, userID uuid.UUID) error {
	if s.repo == nil {
		return ErrCalendarNotConfigured
	}
	connection, err := s.repo.GetActiveConnection(ctx, workspaceID, userID, ProviderGoogle)
	if err != nil {
		return err
	}
	if err := s.syncConnection(ctx, connection); err != nil {
		return err
	}
	return s.enqueueScheduleReconciliation(ctx, connection.UserID)
}

func (s *Service) ListSchedule(ctx context.Context, workspaceID, userID uuid.UUID, startAt, endAt time.Time) (CoreSchedule, error) {
	if s.repo == nil {
		return CoreSchedule{}, ErrCalendarNotConfigured
	}
	if err := validateScheduleRange(startAt, endAt); err != nil {
		return CoreSchedule{}, err
	}
	busyWindows, err := s.repo.ListBusyWindows(ctx, workspaceID, userID, startAt, endAt)
	if err != nil {
		return CoreSchedule{}, err
	}
	blocks, err := s.repo.ListScheduleBlocks(ctx, workspaceID, userID, startAt, endAt)
	if err != nil {
		return CoreSchedule{}, err
	}
	return CoreSchedule{
		StartAt:     startAt.UTC(),
		EndAt:       endAt.UTC(),
		Timezone:    s.scheduleTimezone(ctx, workspaceID, userID),
		BusyWindows: busyWindows,
		Blocks:      blocks,
	}, nil
}

// ListSchedulingAvailability is an internal planning view. It includes the
// user's blocks across workspaces but the repository redacts other-workspace
// story details so account-wide collision protection cannot leak content.
func (s *Service) ListSchedulingAvailability(ctx context.Context, workspaceID, userID uuid.UUID, startAt, endAt time.Time) (CoreSchedule, error) {
	if s.repo == nil {
		return CoreSchedule{}, ErrCalendarNotConfigured
	}
	if err := validateScheduleRange(startAt, endAt); err != nil {
		return CoreSchedule{}, err
	}
	busyWindows, err := s.repo.ListBusyWindows(ctx, workspaceID, userID, startAt, endAt)
	if err != nil {
		return CoreSchedule{}, err
	}
	scheduleRepo, err := s.scheduleReconciliationRepository()
	if err != nil {
		return CoreSchedule{}, err
	}
	blocks, err := scheduleRepo.ListSchedulingBlocksForUser(ctx, workspaceID, userID, startAt, endAt)
	if err != nil {
		return CoreSchedule{}, err
	}
	return CoreSchedule{
		StartAt: startAt.UTC(), EndAt: endAt.UTC(), Timezone: s.scheduleTimezone(ctx, workspaceID, userID),
		BusyWindows: busyWindows, Blocks: blocks,
	}, nil
}

func (s *Service) ListManualSchedulePreference(ctx context.Context, workspaceID, userID uuid.UUID) (CoreSchedulePreference, error) {
	feedbackRepo, ok := s.repo.(ScheduleFeedbackRepository)
	if !ok {
		return CoreSchedulePreference{}, nil
	}
	events, err := feedbackRepo.ListManualScheduleRescheduleEvents(ctx, workspaceID, userID, time.Now().UTC().Add(-90*24*time.Hour))
	if err != nil {
		return CoreSchedulePreference{}, err
	}
	if len(events) == 0 {
		return CoreSchedulePreference{}, nil
	}

	now := time.Now().UTC()
	var weightedStart float64
	var totalWeight float64
	for _, event := range events {
		location, locationErr := time.LoadLocation(fallbackTimezone(event.Timezone))
		if locationErr != nil {
			location = time.UTC
		}
		localStart := event.NextStartAt.In(location)
		minutes := localStart.Hour()*60 + localStart.Minute()
		ageDays := now.Sub(event.CreatedAt.UTC()).Hours() / 24
		if ageDays < 0 {
			ageDays = 0
		}
		weight := 1 / (1 + ageDays/30)
		weightedStart += float64(minutes) * weight
		totalWeight += weight
	}
	if totalWeight == 0 {
		return CoreSchedulePreference{}, nil
	}
	preferredStartMinute := int(weightedStart/totalWeight + 0.5)
	return CoreSchedulePreference{
		PreferredStartMinute: &preferredStartMinute,
		SampleCount:          len(events),
		Confidence:           minFloat(totalWeight/3, 1),
	}, nil
}

func minFloat(left, right float64) float64 {
	if left < right {
		return left
	}
	return right
}

func (s *Service) scheduleTimezone(ctx context.Context, workspaceID, userID uuid.UUID) string {
	connection, err := s.repo.GetActiveConnection(ctx, workspaceID, userID, ProviderGoogle)
	if err != nil {
		return "UTC"
	}
	return fallbackTimezone(connection.Timezone)
}

func (s *Service) ListCalendarView(ctx context.Context, workspaceID, userID uuid.UUID, startAt, endAt time.Time) (CoreCalendarView, error) {
	if s.repo == nil {
		return CoreCalendarView{}, ErrCalendarNotConfigured
	}
	if err := validateScheduleRange(startAt, endAt); err != nil {
		return CoreCalendarView{}, err
	}
	events, err := s.repo.ListCalendarEvents(ctx, workspaceID, userID, startAt, endAt)
	if err != nil {
		return CoreCalendarView{}, err
	}
	busyWindows, err := s.repo.ListBusyWindows(ctx, workspaceID, userID, startAt, endAt)
	if err != nil {
		return CoreCalendarView{}, err
	}
	scheduleRepo, err := s.scheduleReconciliationRepository()
	if err != nil {
		return CoreCalendarView{}, err
	}
	blocks, err := scheduleRepo.ListSchedulingBlocksForUser(ctx, workspaceID, userID, startAt, endAt)
	if err != nil {
		return CoreCalendarView{}, err
	}
	scheduleIssues := []CoreScheduleIssue{}
	if issueRepo, ok := s.repo.(ScheduleIssueRepository); ok {
		scheduleIssues, err = issueRepo.ListScheduleIssues(ctx, workspaceID, userID)
		if err != nil {
			return CoreCalendarView{}, err
		}
	}
	return CoreCalendarView{
		StartAt:        startAt.UTC(),
		EndAt:          endAt.UTC(),
		Events:         events,
		BusyWindows:    busyWindows,
		Blocks:         blocks,
		ScheduleIssues: scheduleIssues,
	}, nil
}

func (s *Service) GetCalendarEvent(ctx context.Context, workspaceID, userID, eventID uuid.UUID) (CoreCalendarEvent, error) {
	if s.repo == nil {
		return CoreCalendarEvent{}, ErrCalendarNotConfigured
	}
	if eventID == uuid.Nil {
		return CoreCalendarEvent{}, ErrCalendarEventNotFound
	}
	return s.repo.GetCalendarEvent(ctx, workspaceID, userID, eventID)
}

func (s *Service) CreateScheduleBlock(ctx context.Context, input CoreScheduleBlockInput) (CoreScheduleBlock, error) {
	if s.repo == nil {
		return CoreScheduleBlock{}, ErrCalendarNotConfigured
	}
	normalized, err := normalizeScheduleBlockInput(input, s.now())
	if err != nil {
		return CoreScheduleBlock{}, err
	}
	if err := s.validateScheduleStory(ctx, normalized); err != nil {
		return CoreScheduleBlock{}, err
	}
	return s.repo.CreateScheduleBlock(ctx, normalized)
}

func (s *Service) UpdateScheduleBlock(ctx context.Context, input CoreScheduleBlockInput) (CoreScheduleBlock, error) {
	if s.repo == nil {
		return CoreScheduleBlock{}, ErrCalendarNotConfigured
	}
	if input.ID == uuid.Nil {
		return CoreScheduleBlock{}, ErrInvalidScheduleBlock
	}
	existing, err := s.repo.GetScheduleBlock(ctx, input.WorkspaceID, input.UserID, input.ID)
	if err != nil {
		return CoreScheduleBlock{}, err
	}
	if existing.Source == ScheduleBlockSourceMaya {
		return CoreScheduleBlock{}, ErrManagedScheduleBlock
	}
	normalized, err := normalizeScheduleBlockInput(input, s.now())
	if err != nil {
		return CoreScheduleBlock{}, err
	}
	if err := s.validateScheduleStory(ctx, normalized); err != nil {
		return CoreScheduleBlock{}, err
	}
	normalized.ID = input.ID
	return s.repo.UpdateScheduleBlock(ctx, normalized)
}

func (s *Service) ManuallyRescheduleScheduleBlock(ctx context.Context, input ManualScheduleBlockInput) (CoreScheduleBlock, error) {
	if s.repo == nil {
		return CoreScheduleBlock{}, ErrCalendarNotConfigured
	}
	if input.WorkspaceID == uuid.Nil || input.UserID == uuid.Nil || input.ActorID == uuid.Nil || input.BlockID == uuid.Nil || input.ClientMutationID == uuid.Nil {
		return CoreScheduleBlock{}, ErrInvalidScheduleBlock
	}
	if input.Change != ManualScheduleBlockChangeMove && input.Change != ManualScheduleBlockChangeResize {
		return CoreScheduleBlock{}, ErrInvalidScheduleBlock
	}
	if err := validateScheduleRange(input.StartAt, input.EndAt); err != nil {
		return CoreScheduleBlock{}, err
	}
	if strings.TrimSpace(input.Timezone) == "" {
		input.Timezone = "UTC"
	}
	manualRepo, ok := s.repo.(ManualScheduleBlockRepository)
	if !ok {
		return CoreScheduleBlock{}, ErrCalendarNotConfigured
	}
	block, err := manualRepo.ManuallyRescheduleScheduleBlock(ctx, input)
	if err != nil {
		return CoreScheduleBlock{}, err
	}
	if s.cfg.Updates != nil {
		if publishErr := s.cfg.Updates.PublishCalendarUpdated(ctx, input.WorkspaceID, input.UserID, uuid.Nil, s.now().UTC()); publishErr != nil && s.log != nil {
			s.log.Error(ctx, "failed to publish manual calendar update", "error", publishErr, "block_id", input.BlockID)
		}
	}
	return block, nil
}

func (s *Service) DeleteScheduleBlock(ctx context.Context, workspaceID, userID, blockID uuid.UUID) error {
	if s.repo == nil {
		return ErrCalendarNotConfigured
	}
	if blockID == uuid.Nil {
		return ErrInvalidScheduleBlock
	}
	existing, err := s.repo.GetScheduleBlock(ctx, workspaceID, userID, blockID)
	if err != nil {
		return err
	}
	if existing.Source == ScheduleBlockSourceMaya {
		return ErrManagedScheduleBlock
	}
	return s.repo.DeleteScheduleBlock(ctx, workspaceID, userID, blockID)
}

func (s *Service) ReconcileMayaScheduleBlocks(ctx context.Context, input MayaScheduleReconcileInput) (CoreScheduleReconcileResult, error) {
	if s.repo == nil {
		return CoreScheduleReconcileResult{}, ErrCalendarNotConfigured
	}
	if input.WorkspaceID == uuid.Nil || input.UserID == uuid.Nil || input.StoryID == uuid.Nil {
		return CoreScheduleReconcileResult{}, ErrInvalidScheduleBlock
	}
	if input.AllowConflicts && !input.Locked {
		return CoreScheduleReconcileResult{}, ErrInvalidScheduleBlock
	}
	if len(input.Segments) > 0 {
		if exists, err := s.repo.ScheduleStoryExists(ctx, input.WorkspaceID, input.UserID, input.StoryID); err != nil {
			return CoreScheduleReconcileResult{}, err
		} else if !exists {
			return CoreScheduleReconcileResult{}, ErrCalendarAccessDenied
		}
	}
	for index := range input.Segments {
		input.Segments[index].Title = strings.TrimSpace(input.Segments[index].Title)
		if input.Segments[index].Title == "" || len(input.Segments[index].Title) > 255 {
			return CoreScheduleReconcileResult{}, ErrInvalidScheduleBlock
		}
		if err := validateScheduleRange(input.Segments[index].StartAt, input.Segments[index].EndAt); err != nil {
			return CoreScheduleReconcileResult{}, err
		}
	}
	scheduleRepo, err := s.scheduleReconciliationRepository()
	if err != nil {
		return CoreScheduleReconcileResult{}, err
	}
	result, err := scheduleRepo.ReconcileMayaScheduleBlocks(ctx, input)
	if err != nil {
		return CoreScheduleReconcileResult{}, err
	}
	if s.cfg.Updates != nil && scheduleReconcileChangesCalendar(result.Actions) {
		if publishErr := s.cfg.Updates.PublishCalendarUpdated(
			ctx,
			input.WorkspaceID,
			input.UserID,
			uuid.Nil,
			s.now().UTC(),
		); publishErr != nil {
			// Reconciliation is already durably committed. Realtime invalidation is
			// advisory and must not make the caller compensate a successful plan.
			if s.log != nil {
				s.log.Error(ctx, "failed to publish reconciled calendar update", "error", publishErr, "story_id", input.StoryID, "user_id", input.UserID)
			}
		}
	}
	return result, nil
}

func scheduleReconcileChangesCalendar(actions []ScheduleReconcileAction) bool {
	for _, action := range actions {
		if action != ScheduleReconcileActionUnchanged {
			return true
		}
	}
	return false
}

func (s *Service) ListMayaScheduleBlocksForStory(ctx context.Context, workspaceID, userID, storyID uuid.UUID) ([]CoreScheduleBlock, error) {
	if workspaceID == uuid.Nil || userID == uuid.Nil || storyID == uuid.Nil {
		return nil, ErrInvalidScheduleBlock
	}
	scheduleRepo, err := s.scheduleReconciliationRepository()
	if err != nil {
		return nil, err
	}
	return scheduleRepo.ListMayaScheduleBlocksForStory(ctx, workspaceID, userID, storyID)
}

func (s *Service) MayaScheduleOwnershipExists(ctx context.Context, workspaceID, userID, storyID uuid.UUID) (bool, error) {
	if workspaceID == uuid.Nil || userID == uuid.Nil || storyID == uuid.Nil {
		return false, ErrInvalidScheduleBlock
	}
	scheduleRepo, err := s.scheduleReconciliationRepository()
	if err != nil {
		return false, err
	}
	return scheduleRepo.MayaScheduleOwnershipExists(ctx, workspaceID, userID, storyID)
}

func (s *Service) DispatchScheduleEventOutbox(ctx context.Context, userID uuid.UUID) error {
	if s.repo == nil || userID == uuid.Nil {
		return ErrCalendarNotConfigured
	}
	scheduleRepo, err := s.scheduleReconciliationRepository()
	if err != nil {
		return err
	}
	var dispatchErr error
	lockErr := scheduleRepo.WithScheduleEventDispatchLock(ctx, userID, func(outbox ScheduleEventOutboxStore) error {
		connection, cleanupPending, err := s.repo.GetScheduleEventDispatchConnection(ctx, userID)
		if err != nil {
			if errors.Is(err, ErrCalendarNotFound) {
				return nil
			}
			return err
		}
		if cleanupPending && !connection.CanDeleteOwnedEvents() {
			return terminallyFinalizeScheduleCleanup(
				ctx,
				outbox,
				userID,
				"Calendar cleanup could not call Google because the connection never granted owned-event access.",
			)
		}
		if !cleanupPending && !connection.CanWriteEvents() {
			return nil
		}
		provider, err := s.provider(connection.Provider)
		if err != nil {
			if cleanupPending {
				return terminallyFinalizeScheduleCleanup(ctx, outbox, userID, "Calendar cleanup could not initialize the provider writer.")
			}
			return err
		}
		eventWriter, ok := provider.(CalendarEventWriter)
		if !ok {
			if cleanupPending {
				return terminallyFinalizeScheduleCleanup(ctx, outbox, userID, "Calendar cleanup provider does not support event deletion.")
			}
			return ErrCalendarNotConfigured
		}
		token, err := s.decryptTokenPayload(connection.TokenPayload)
		if err != nil {
			if cleanupPending {
				return terminallyFinalizeScheduleCleanup(ctx, outbox, userID, "Calendar cleanup credentials could not be decrypted.")
			}
			return err
		}
		for {
			items, err := outbox.ListPendingScheduleEventOutbox(ctx, userID, 100)
			if err != nil {
				return err
			}
			if len(items) == 0 {
				return outbox.DeleteCleanupPendingConnectionIfDrained(ctx, userID)
			}
			for index, item := range items {
				var event ExternalScheduleEventInput
				isMember := false
				if !cleanupPending {
					var membershipErr error
					isMember, membershipErr = s.repo.WorkspaceMemberExists(ctx, item.WorkspaceID, userID)
					if membershipErr != nil {
						return membershipErr
					}
				}
				needsEventPayload := item.Operation == ScheduleEventOperationUpsert && !cleanupPending && isMember
				if needsEventPayload {
					if err := json.Unmarshal(item.Payload, &event); err != nil {
						if markErr := outbox.MarkScheduleEventOutboxFailed(ctx, item, "Calendar event payload could not be decoded.", true); markErr != nil {
							return errors.Join(err, markErr)
						}
						if releaseErr := outbox.ReleaseScheduleEventOutbox(ctx, scheduleOutboxIDs(items[index+1:])); releaseErr != nil {
							return errors.Join(err, releaseErr)
						}
						if cleanupPending {
							if finalizeErr := outbox.DeleteCleanupPendingConnectionIfDrained(ctx, userID); finalizeErr != nil {
								return errors.Join(err, finalizeErr)
							}
						}
						dispatchErr = fmt.Errorf("decode calendar schedule event outbox: %w", err)
						return nil
					}
				}
				forceDelete := cleanupPending || item.Operation == ScheduleEventOperationDelete || item.Operation == ScheduleEventOperationUpsert && !isMember
				if !forceDelete && item.Operation == ScheduleEventOperationUpsert {
					current, currentErr := outbox.ScheduleEventUpsertIsCurrent(ctx, item, event)
					if currentErr != nil {
						return currentErr
					}
					forceDelete = !current
				}
				var writeErr error
				failureMessage := "Calendar event update failed."
				if forceDelete {
					writeErr = eventWriter.DeleteScheduleEvent(ctx, token, item.CalendarID, item.ProviderEventID)
					failureMessage = "Calendar event cleanup failed."
				} else {
					writeErr = eventWriter.UpsertScheduleEvent(ctx, token, event)
				}
				if writeErr != nil {
					failureMessage = fmt.Sprintf("%s %v", failureMessage, writeErr)
					terminal := isPermanentCalendarWriteError(writeErr) || item.AttemptCount >= maximumScheduleEventOutboxAttempts
					if markErr := outbox.MarkScheduleEventOutboxFailed(ctx, item, failureMessage, terminal); markErr != nil {
						return errors.Join(writeErr, markErr)
					}
					if releaseErr := outbox.ReleaseScheduleEventOutbox(ctx, scheduleOutboxIDs(items[index+1:])); releaseErr != nil {
						return errors.Join(writeErr, releaseErr)
					}
					if terminal && cleanupPending {
						if finalizeErr := outbox.DeleteCleanupPendingConnectionIfDrained(ctx, userID); finalizeErr != nil {
							return errors.Join(writeErr, finalizeErr)
						}
					}
					dispatchErr = writeErr
					return nil
				}
				processedItem := item
				if forceDelete {
					processedItem.Operation = ScheduleEventOperationDelete
				}
				if err := outbox.MarkScheduleEventOutboxProcessed(ctx, processedItem, ScheduleEventSyncHash(event)); err != nil {
					return err
				}
			}
		}
	})
	return errors.Join(lockErr, dispatchErr)
}

func terminallyFinalizeScheduleCleanup(ctx context.Context, outbox ScheduleEventOutboxStore, userID uuid.UUID, message string) error {
	for {
		items, err := outbox.ListPendingScheduleEventOutbox(ctx, userID, 100)
		if err != nil {
			return err
		}
		if len(items) == 0 {
			return outbox.DeleteCleanupPendingConnectionIfDrained(ctx, userID)
		}
		for _, item := range items {
			if err := outbox.MarkScheduleEventOutboxFailed(ctx, item, message, true); err != nil {
				return err
			}
		}
	}
}

func isPermanentCalendarWriteError(err error) bool {
	var apiErr *googleapi.Error
	if !errors.As(err, &apiErr) {
		return false
	}
	if apiErr.Code < 400 || apiErr.Code >= 500 {
		return false
	}
	switch apiErr.Code {
	case 408, 409, 429:
		return false
	}
	for _, item := range apiErr.Errors {
		switch item.Reason {
		case "rateLimitExceeded", "userRateLimitExceeded", "backendError":
			return false
		}
	}
	return true
}

func (s *Service) DispatchReadyScheduleEventOutbox(ctx context.Context) (int, error) {
	if s.repo == nil {
		return 0, ErrCalendarNotConfigured
	}
	scheduleRepo, err := s.scheduleReconciliationRepository()
	if err != nil {
		return 0, err
	}
	userIDs, err := scheduleRepo.ListReadyScheduleEventOutboxUsers(ctx, 100)
	if err != nil {
		return 0, err
	}
	var dispatchErr error
	for _, userID := range userIDs {
		if err := s.DispatchScheduleEventOutbox(ctx, userID); err != nil {
			dispatchErr = errors.Join(dispatchErr, err)
		}
	}
	return len(userIDs), dispatchErr
}

func scheduleOutboxIDs(items []CoreScheduleEventOutbox) []uuid.UUID {
	ids := make([]uuid.UUID, 0, len(items))
	for _, item := range items {
		if item.ID != uuid.Nil {
			ids = append(ids, item.ID)
		}
	}
	return ids
}

func (s *Service) scheduleReconciliationRepository() (ScheduleReconciliationRepository, error) {
	repo, ok := s.repo.(ScheduleReconciliationRepository)
	if !ok {
		return nil, ErrCalendarNotConfigured
	}
	return repo, nil
}

func (s *Service) validateScheduleStory(ctx context.Context, input CoreScheduleBlockInput) error {
	if input.StoryID == nil {
		return nil
	}
	exists, err := s.repo.ScheduleStoryExists(ctx, input.WorkspaceID, input.UserID, *input.StoryID)
	if err != nil {
		return err
	}
	if !exists {
		return ErrInvalidScheduleBlock
	}
	return nil
}

func (s *Service) syncConnection(ctx context.Context, connection CoreConnection) error {
	connection, err := s.repo.BeginConnectionSync(ctx, connection)
	if err != nil {
		return err
	}
	provider, err := s.provider(connection.Provider)
	if err != nil {
		return s.failConnectionSync(ctx, connection, err)
	}
	token, err := s.decryptTokenPayload(connection.TokenPayload)
	if err != nil {
		return s.failConnectionSync(ctx, connection, err)
	}
	timeMin := s.now().Add(defaultSyncLookback)
	timeMax := s.now().Add(defaultSyncLookahead)
	snapshot, err := provider.SyncCalendar(ctx, token, BusyWindowInput{
		ConnectionID: connection.ID,
		WorkspaceID:  connection.WorkspaceID,
		UserID:       connection.UserID,
		TimeMin:      timeMin,
		TimeMax:      timeMax,
		Timezone:     fallbackTimezone(connection.Timezone),
	})
	if err != nil {
		return s.failConnectionSync(ctx, connection, err)
	}
	for i := range snapshot.Events {
		snapshot.Events[i].ConnectionID = connection.ID
		snapshot.Events[i].WorkspaceID = connection.WorkspaceID
		snapshot.Events[i].UserID = connection.UserID
		snapshot.Events[i].Provider = connection.Provider
		if strings.TrimSpace(snapshot.Events[i].CalendarID) == "" {
			snapshot.Events[i].CalendarID = "primary"
		}
		if strings.TrimSpace(snapshot.Events[i].Visibility) == "" {
			snapshot.Events[i].Visibility = "default"
		}
		if snapshot.Events[i].Attendees == nil {
			snapshot.Events[i].Attendees = []CoreCalendarParticipant{}
		}
	}
	for i := range snapshot.BusyWindows {
		snapshot.BusyWindows[i].ConnectionID = connection.ID
		snapshot.BusyWindows[i].WorkspaceID = connection.WorkspaceID
		snapshot.BusyWindows[i].UserID = connection.UserID
		snapshot.BusyWindows[i].Provider = connection.Provider
	}
	if err := s.repo.ReplaceCalendarSnapshot(ctx, connection, snapshot); err != nil {
		if errors.Is(err, ErrCalendarSyncSuperseded) {
			return err
		}
		return s.failConnectionSync(ctx, connection, err)
	}
	return s.completeConnectionSync(ctx, connection)
}

func (s *Service) completeConnectionSync(ctx context.Context, connection CoreConnection) error {
	syncedAt := s.now().UTC()
	if err := s.repo.MarkConnectionSynced(
		ctx,
		connection.WorkspaceID,
		connection.ID,
		connection.CredentialGeneration,
		syncedAt,
	); err != nil {
		return err
	}
	if s.cfg.Updates == nil {
		return nil
	}
	if err := s.cfg.Updates.PublishCalendarUpdated(
		ctx,
		connection.WorkspaceID,
		connection.UserID,
		connection.ID,
		syncedAt,
	); err != nil {
		return fmt.Errorf("publish calendar update: %w", err)
	}
	return nil
}

func (s *Service) replaceNotificationChannel(ctx context.Context, connection CoreConnection) error {
	address := strings.TrimSpace(s.cfg.WebhookURL)
	if address == "" {
		if !s.cfg.RequireWebhook {
			return nil
		}
		return ErrCalendarWebhookNotConfigured
	}
	parsedAddress, err := url.Parse(address)
	if err != nil || parsedAddress.Scheme != "https" || parsedAddress.Host == "" {
		return fmt.Errorf("calendar webhook must be a public HTTPS URL")
	}
	provider, err := s.provider(connection.Provider)
	if err != nil {
		return err
	}
	token, err := s.decryptTokenPayload(connection.TokenPayload)
	if err != nil {
		return err
	}
	channelID := uuid.NewString()
	channel, err := provider.WatchCalendar(ctx, token, CalendarWatchInput{
		ChannelID: channelID,
		Address:   parsedAddress.String(),
		Token:     s.notificationToken(connection.ID, channelID),
		TTL:       googleWatchTTL,
	})
	if err != nil {
		return err
	}
	if strings.TrimSpace(channel.ChannelID) == "" ||
		strings.TrimSpace(channel.ResourceID) == "" ||
		channel.ExpiresAt.IsZero() {
		return errors.New("calendar provider returned an incomplete notification channel")
	}
	if err := s.repo.SetNotificationChannel(ctx, connection, channel); err != nil {
		_ = provider.StopCalendarWatch(ctx, token, channel)
		return err
	}
	if connection.NotificationChannelID != "" && connection.NotificationResourceID != "" {
		old := CalendarWatchChannel{
			ChannelID:  connection.NotificationChannelID,
			ResourceID: connection.NotificationResourceID,
		}
		if err := provider.StopCalendarWatch(ctx, token, old); err != nil && s.log != nil {
			s.log.Error(ctx, "failed to stop replaced calendar notification channel", "err", err, "connection_id", connection.ID)
		}
	}
	return nil
}

func (s *Service) stopNotificationChannel(ctx context.Context, connection CoreConnection) {
	if connection.NotificationChannelID == "" || connection.NotificationResourceID == "" {
		return
	}
	provider, err := s.provider(connection.Provider)
	if err != nil {
		return
	}
	token, err := s.decryptTokenPayload(connection.TokenPayload)
	if err != nil {
		return
	}
	err = provider.StopCalendarWatch(ctx, token, CalendarWatchChannel{
		ChannelID:  connection.NotificationChannelID,
		ResourceID: connection.NotificationResourceID,
	})
	if err != nil && s.log != nil {
		s.log.Error(ctx, "failed to stop calendar notification channel", "err", err, "connection_id", connection.ID)
	}
}

func (s *Service) notificationToken(connectionID uuid.UUID, channelID string) string {
	payload := connectionID.String() + "." + channelID
	mac := hmac.New(sha256.New, []byte(s.cfg.SecretKey))
	_, _ = mac.Write([]byte(payload))
	return payload + "." + base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}

func (s *Service) verifyNotificationToken(value string) (uuid.UUID, string, error) {
	parts := strings.Split(strings.TrimSpace(value), ".")
	if len(parts) != 3 {
		return uuid.Nil, "", ErrInvalidCalendarNotification
	}
	connectionID, err := uuid.Parse(parts[0])
	if err != nil {
		return uuid.Nil, "", ErrInvalidCalendarNotification
	}
	expected := s.notificationToken(connectionID, parts[1])
	if !hmac.Equal([]byte(expected), []byte(strings.TrimSpace(value))) {
		return uuid.Nil, "", ErrInvalidCalendarNotification
	}
	return connectionID, parts[1], nil
}

func (s *Service) normalizeDelta(connection CoreConnection, delta *CalendarSyncDelta) {
	for i := range delta.Events {
		delta.Events[i].ConnectionID = connection.ID
		delta.Events[i].WorkspaceID = connection.WorkspaceID
		delta.Events[i].UserID = connection.UserID
		delta.Events[i].Provider = connection.Provider
		if strings.TrimSpace(delta.Events[i].CalendarID) == "" {
			delta.Events[i].CalendarID = "primary"
		}
		if strings.TrimSpace(delta.Events[i].Visibility) == "" {
			delta.Events[i].Visibility = "default"
		}
		if delta.Events[i].Attendees == nil {
			delta.Events[i].Attendees = []CoreCalendarParticipant{}
		}
	}
	for i := range delta.BusyWindows {
		delta.BusyWindows[i].ConnectionID = connection.ID
		delta.BusyWindows[i].WorkspaceID = connection.WorkspaceID
		delta.BusyWindows[i].UserID = connection.UserID
		delta.BusyWindows[i].Provider = connection.Provider
	}
}

func (s *Service) failConnectionSync(ctx context.Context, connection CoreConnection, syncErr error) error {
	if s.log != nil {
		s.log.Error(ctx, "calendar sync failed", "err", syncErr, "connection_id", connection.ID)
	}
	markErr := s.repo.MarkConnectionSyncFailed(
		ctx,
		connection.WorkspaceID,
		connection.ID,
		connection.CredentialGeneration,
		"Calendar could not be refreshed.",
	)
	if markErr == nil {
		return syncErr
	}
	if errors.Is(markErr, ErrCalendarSyncSuperseded) {
		return markErr
	}
	return errors.Join(syncErr, markErr)
}

func (s *Service) provider(provider Provider) (CalendarProvider, error) {
	if s.cfg.Providers == nil {
		return nil, ErrCalendarNotConfigured
	}
	configured := s.cfg.Providers[provider]
	if configured == nil {
		return nil, ErrCalendarNotConfigured
	}
	return configured, nil
}

func (s *Service) signState(claims stateClaims) (string, error) {
	if strings.TrimSpace(s.cfg.SecretKey) == "" {
		return "", ErrCalendarNotConfigured
	}
	payloadBytes, err := json.Marshal(claims)
	if err != nil {
		return "", err
	}
	payload := base64.RawURLEncoding.EncodeToString(payloadBytes)
	mac := hmac.New(sha256.New, []byte(s.cfg.SecretKey))
	_, _ = mac.Write([]byte(payload))
	sig := hex.EncodeToString(mac.Sum(nil))
	return payload + "." + sig, nil
}

func (s *Service) verifyState(value string) (stateClaims, error) {
	if strings.TrimSpace(s.cfg.SecretKey) == "" {
		return stateClaims{}, ErrCalendarNotConfigured
	}
	parts := strings.Split(value, ".")
	if len(parts) != 2 {
		return stateClaims{}, ErrInvalidCalendarState
	}
	mac := hmac.New(sha256.New, []byte(s.cfg.SecretKey))
	_, _ = mac.Write([]byte(parts[0]))
	expected := hex.EncodeToString(mac.Sum(nil))
	if !hmac.Equal([]byte(expected), []byte(parts[1])) {
		return stateClaims{}, ErrInvalidCalendarState
	}
	payloadBytes, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return stateClaims{}, err
	}
	var claims stateClaims
	if err := json.Unmarshal(payloadBytes, &claims); err != nil {
		return stateClaims{}, err
	}
	return claims, nil
}

func (s *Service) verifyActiveState(value string) (stateClaims, error) {
	claims, err := s.verifyState(strings.TrimSpace(value))
	if err != nil {
		return stateClaims{}, err
	}
	if claims.WorkspaceID == uuid.Nil ||
		claims.UserID == uuid.Nil ||
		claims.Provider != ProviderGoogle ||
		strings.TrimSpace(claims.WorkspaceSlug) == "" {
		return stateClaims{}, ErrInvalidCalendarState
	}
	if !s.now().Before(time.Unix(claims.ExpiresAt, 0)) {
		return stateClaims{}, fmt.Errorf("%w: expired", ErrInvalidCalendarState)
	}
	return claims, nil
}

func (s *Service) encryptTokenPayload(token ProviderToken) (string, error) {
	key := s.encryptionKey()
	block, err := aes.NewCipher(key[:])
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(randReader{s.randBytes}, nonce); err != nil {
		return "", err
	}
	plaintext, err := json.Marshal(token)
	if err != nil {
		return "", err
	}
	ciphertext := gcm.Seal(nil, nonce, plaintext, nil)
	return base64.RawURLEncoding.EncodeToString(append(nonce, ciphertext...)), nil
}

func (s *Service) decryptTokenPayload(value string) (ProviderToken, error) {
	raw, err := base64.RawURLEncoding.DecodeString(strings.TrimSpace(value))
	if err != nil {
		return ProviderToken{}, err
	}
	key := s.encryptionKey()
	block, err := aes.NewCipher(key[:])
	if err != nil {
		return ProviderToken{}, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return ProviderToken{}, err
	}
	if len(raw) < gcm.NonceSize() {
		return ProviderToken{}, errors.New("calendar token payload is too short")
	}
	nonce := raw[:gcm.NonceSize()]
	ciphertext := raw[gcm.NonceSize():]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return ProviderToken{}, err
	}
	var token ProviderToken
	if err := json.Unmarshal(plaintext, &token); err != nil {
		return ProviderToken{}, err
	}
	return token, nil
}

func (s *Service) encryptionKey() [32]byte {
	return sha256.Sum256([]byte(s.cfg.SecretKey))
}

func (s *Service) workspaceCalendarURL(workspaceSlug, query string) string {
	base := strings.TrimRight(s.cfg.WebsiteURL, "/")
	if base == "" {
		base = "/"
	}
	path := fmt.Sprintf("%s/%s/settings/account/calendar", base, url.PathEscape(workspaceSlug))
	if strings.TrimSpace(query) == "" {
		return path
	}
	return path + "?" + query
}

func fallbackTimezone(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return "UTC"
	}
	return value
}

func validateScheduleRange(startAt, endAt time.Time) error {
	if startAt.IsZero() || endAt.IsZero() || !endAt.After(startAt) {
		return ErrInvalidScheduleRange
	}
	if endAt.Sub(startAt) > 93*24*time.Hour {
		return ErrInvalidScheduleRange
	}
	return nil
}

func normalizeScheduleBlockInput(input CoreScheduleBlockInput, now time.Time) (CoreScheduleBlockInput, error) {
	if input.WorkspaceID == uuid.Nil || input.UserID == uuid.Nil {
		return CoreScheduleBlockInput{}, ErrInvalidScheduleBlock
	}
	if err := validateScheduleRange(input.StartAt, input.EndAt); err != nil {
		return CoreScheduleBlockInput{}, err
	}
	if input.StartAt.Before(now.Add(defaultSyncLookback)) || input.EndAt.After(now.Add(defaultSyncLookahead)) {
		return CoreScheduleBlockInput{}, ErrInvalidScheduleRange
	}
	input.Title = strings.TrimSpace(input.Title)
	if input.Title == "" {
		return CoreScheduleBlockInput{}, ErrInvalidScheduleBlock
	}
	switch input.BlockType {
	case ScheduleBlockTypeWork:
		if input.StoryID == nil || *input.StoryID == uuid.Nil {
			return CoreScheduleBlockInput{}, ErrInvalidScheduleBlock
		}
	case ScheduleBlockTypeFocus:
		input.StoryID = nil
	default:
		return CoreScheduleBlockInput{}, ErrInvalidScheduleBlock
	}
	if input.Source == "" {
		input.Source = ScheduleBlockSourceUser
	}
	switch input.Source {
	case ScheduleBlockSourceUser, ScheduleBlockSourceMaya:
	default:
		return CoreScheduleBlockInput{}, ErrInvalidScheduleBlock
	}
	input.StartAt = input.StartAt.UTC()
	input.EndAt = input.EndAt.UTC()
	return input, nil
}

type randReader struct {
	read func([]byte) (int, error)
}

func (r randReader) Read(p []byte) (int, error) {
	return r.read(p)
}
