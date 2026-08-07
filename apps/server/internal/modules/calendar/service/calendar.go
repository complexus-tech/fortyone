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
)

var (
	ErrCalendarNotConfigured         = errors.New("calendar integration is not configured")
	ErrInvalidCalendarState          = errors.New("invalid calendar setup state")
	ErrCalendarNotFound              = errors.New("calendar connection not found")
	ErrCalendarAccessDenied          = errors.New("calendar access denied")
	ErrCalendarCredentialsIncomplete = errors.New("calendar credentials are incomplete")
	ErrCalendarEventNotFound         = errors.New("calendar event not found")
	ErrCalendarSyncSuperseded        = errors.New("calendar sync was superseded by newer credentials")
	ErrInvalidScheduleRange          = errors.New("calendar schedule range is invalid")
	ErrInvalidScheduleBlock          = errors.New("calendar schedule block is invalid")
	ErrCalendarScheduleConflict      = errors.New("calendar time conflicts with an existing meeting or schedule block")
	ErrCalendarScheduleBlockNotFound = errors.New("calendar schedule block not found")
)

const (
	connectStateTTL      = 10 * time.Minute
	defaultSyncLookback  = -7 * 24 * time.Hour
	defaultSyncLookahead = 90 * 24 * time.Hour
)

type Repository interface {
	ListConnections(ctx context.Context, workspaceID uuid.UUID, userID *uuid.UUID) ([]CoreConnection, error)
	GetOwnedConnection(ctx context.Context, workspaceID, userID, connectionID uuid.UUID) (CoreConnection, error)
	GetActiveConnection(ctx context.Context, workspaceID, userID uuid.UUID, provider Provider) (CoreConnection, error)
	WorkspaceMemberExists(ctx context.Context, workspaceID, userID uuid.UUID) (bool, error)
	UpsertConnection(ctx context.Context, input CoreConnectionUpsert) (CoreConnection, error)
	BeginConnectionSync(ctx context.Context, connection CoreConnection) (CoreConnection, error)
	RevokeConnection(ctx context.Context, workspaceID, userID, connectionID uuid.UUID) error
	ReplaceCalendarSnapshot(ctx context.Context, connection CoreConnection, snapshot CalendarSyncSnapshot) error
	ListCalendarEvents(ctx context.Context, workspaceID, userID uuid.UUID, startAt, endAt time.Time) ([]CoreCalendarEventSummary, error)
	GetCalendarEvent(ctx context.Context, workspaceID, userID, eventID uuid.UUID) (CoreCalendarEvent, error)
	ListBusyWindows(ctx context.Context, workspaceID, userID uuid.UUID, startAt, endAt time.Time) ([]CoreBusyWindow, error)
	ListScheduleBlocks(ctx context.Context, workspaceID, userID uuid.UUID, startAt, endAt time.Time) ([]CoreScheduleBlock, error)
	ScheduleStoryExists(ctx context.Context, workspaceID, userID, storyID uuid.UUID) (bool, error)
	CreateScheduleBlock(ctx context.Context, input CoreScheduleBlockInput) (CoreScheduleBlock, error)
	UpdateScheduleBlock(ctx context.Context, input CoreScheduleBlockInput) (CoreScheduleBlock, error)
	DeleteScheduleBlock(ctx context.Context, workspaceID, userID, blockID uuid.UUID) error
	MarkConnectionSynced(ctx context.Context, workspaceID, connectionID, credentialGeneration uuid.UUID, syncedAt time.Time) error
	MarkConnectionSyncFailed(ctx context.Context, workspaceID, connectionID, credentialGeneration uuid.UUID, message string) error
}

type Config struct {
	SecretKey  string
	WebsiteURL string
	Providers  map[Provider]CalendarProvider
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
	if err := s.syncConnection(ctx, connection); err != nil && s.log != nil {
		s.log.Error(ctx, "failed to sync calendar connection after connect", "err", err, "connection_id", connection.ID)
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
	return s.repo.RevokeConnection(ctx, workspaceID, userID, connectionID)
}

func (s *Service) SyncConnection(ctx context.Context, workspaceID, userID, connectionID uuid.UUID) error {
	if s.repo == nil {
		return ErrCalendarNotConfigured
	}
	connection, err := s.repo.GetOwnedConnection(ctx, workspaceID, userID, connectionID)
	if err != nil {
		return err
	}
	return s.syncConnection(ctx, connection)
}

func (s *Service) SyncActiveGoogleConnection(ctx context.Context, workspaceID, userID uuid.UUID) error {
	if s.repo == nil {
		return ErrCalendarNotConfigured
	}
	connection, err := s.repo.GetActiveConnection(ctx, workspaceID, userID, ProviderGoogle)
	if err != nil {
		return err
	}
	return s.syncConnection(ctx, connection)
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
		BusyWindows: busyWindows,
		Blocks:      blocks,
	}, nil
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
	blocks, err := s.repo.ListScheduleBlocks(ctx, workspaceID, userID, startAt, endAt)
	if err != nil {
		return CoreCalendarView{}, err
	}
	return CoreCalendarView{
		StartAt:     startAt.UTC(),
		EndAt:       endAt.UTC(),
		Events:      events,
		BusyWindows: busyWindows,
		Blocks:      blocks,
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

func (s *Service) DeleteScheduleBlock(ctx context.Context, workspaceID, userID, blockID uuid.UUID) error {
	if s.repo == nil {
		return ErrCalendarNotConfigured
	}
	if blockID == uuid.Nil {
		return ErrInvalidScheduleBlock
	}
	return s.repo.DeleteScheduleBlock(ctx, workspaceID, userID, blockID)
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
	return s.repo.MarkConnectionSynced(
		ctx,
		connection.WorkspaceID,
		connection.ID,
		connection.CredentialGeneration,
		s.now().UTC(),
	)
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
	path := fmt.Sprintf("%s/%s/settings/integrations/calendar", base, url.PathEscape(workspaceSlug))
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
