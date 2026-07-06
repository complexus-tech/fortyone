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
	ErrInvalidScheduleRange          = errors.New("calendar schedule range is invalid")
	ErrInvalidScheduleBlock          = errors.New("calendar schedule block is invalid")
	ErrCalendarScheduleBlockNotFound = errors.New("calendar schedule block not found")
)

const (
	connectStateTTL      = 10 * time.Minute
	defaultSyncLookback  = -7 * 24 * time.Hour
	defaultSyncLookahead = 90 * 24 * time.Hour
)

type Repository interface {
	ListConnections(ctx context.Context, workspaceID uuid.UUID, userID *uuid.UUID) ([]CoreConnection, error)
	GetConnection(ctx context.Context, workspaceID, connectionID uuid.UUID) (CoreConnection, error)
	GetActiveConnection(ctx context.Context, workspaceID, userID uuid.UUID, provider Provider) (CoreConnection, error)
	UpsertConnection(ctx context.Context, input CoreConnectionUpsert) (CoreConnection, error)
	RevokeConnection(ctx context.Context, workspaceID, userID, connectionID uuid.UUID) error
	ReplaceBusyWindows(ctx context.Context, connection CoreConnection, windows []CoreBusyWindow) error
	ListBusyWindows(ctx context.Context, workspaceID, userID uuid.UUID, startAt, endAt time.Time) ([]CoreBusyWindow, error)
	ListScheduleBlocks(ctx context.Context, workspaceID, userID uuid.UUID, startAt, endAt time.Time) ([]CoreScheduleBlock, error)
	CreateScheduleBlock(ctx context.Context, input CoreScheduleBlockInput) (CoreScheduleBlock, error)
	UpdateScheduleBlock(ctx context.Context, input CoreScheduleBlockInput) (CoreScheduleBlock, error)
	DeleteScheduleBlock(ctx context.Context, workspaceID, userID, blockID uuid.UUID) error
	MarkConnectionSynced(ctx context.Context, workspaceID, connectionID uuid.UUID, syncedAt time.Time) error
	MarkConnectionSyncFailed(ctx context.Context, workspaceID, connectionID uuid.UUID, message string) error
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

func (s *Service) CompleteConnect(ctx context.Context, code, state string) (CoreConnection, string, error) {
	if s.repo == nil {
		return CoreConnection{}, "", ErrCalendarNotConfigured
	}
	claims, err := s.verifyState(strings.TrimSpace(state))
	if err != nil {
		return CoreConnection{}, "", err
	}
	if !s.now().Before(time.Unix(claims.ExpiresAt, 0)) {
		return CoreConnection{}, "", fmt.Errorf("%w: expired", ErrInvalidCalendarState)
	}
	provider, err := s.provider(claims.Provider)
	if err != nil {
		return CoreConnection{}, "", err
	}
	token, err := provider.ExchangeCode(ctx, strings.TrimSpace(code))
	if err != nil {
		return CoreConnection{}, "", err
	}
	payload, err := s.encryptTokenPayload(token)
	if err != nil {
		return CoreConnection{}, "", err
	}
	connection, err := s.repo.UpsertConnection(ctx, CoreConnectionUpsert{
		WorkspaceID:    claims.WorkspaceID,
		UserID:         claims.UserID,
		Provider:       claims.Provider,
		ConnectedEmail: strings.TrimSpace(token.ConnectedEmail),
		Timezone:       fallbackTimezone(token.Timezone),
		TokenPayload:   payload,
		Scopes:         token.Scopes,
	})
	if err != nil {
		return CoreConnection{}, "", err
	}
	if err := s.syncConnection(ctx, connection); err != nil && s.log != nil {
		s.log.Error(ctx, "failed to sync calendar connection after connect", "err", err, "connection_id", connection.ID)
	}
	return connection, s.workspaceCalendarURL(claims.WorkspaceSlug, "connected=1"), nil
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
	connection, err := s.repo.GetConnection(ctx, workspaceID, connectionID)
	if err != nil {
		return err
	}
	if connection.UserID != userID {
		return ErrCalendarNotFound
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

func (s *Service) CreateScheduleBlock(ctx context.Context, input CoreScheduleBlockInput) (CoreScheduleBlock, error) {
	if s.repo == nil {
		return CoreScheduleBlock{}, ErrCalendarNotConfigured
	}
	normalized, err := normalizeScheduleBlockInput(input)
	if err != nil {
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
	normalized, err := normalizeScheduleBlockInput(input)
	if err != nil {
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

func (s *Service) syncConnection(ctx context.Context, connection CoreConnection) error {
	provider, err := s.provider(connection.Provider)
	if err != nil {
		return err
	}
	token, err := s.decryptTokenPayload(connection.TokenPayload)
	if err != nil {
		return err
	}
	timeMin := s.now().Add(defaultSyncLookback)
	timeMax := s.now().Add(defaultSyncLookahead)
	windows, err := provider.ListBusyWindows(ctx, token, BusyWindowInput{
		ConnectionID: connection.ID,
		WorkspaceID:  connection.WorkspaceID,
		UserID:       connection.UserID,
		TimeMin:      timeMin,
		TimeMax:      timeMax,
		Timezone:     fallbackTimezone(connection.Timezone),
	})
	if err != nil {
		message := err.Error()
		_ = s.repo.MarkConnectionSyncFailed(ctx, connection.WorkspaceID, connection.ID, message)
		return err
	}
	for i := range windows {
		windows[i].ConnectionID = connection.ID
		windows[i].WorkspaceID = connection.WorkspaceID
		windows[i].UserID = connection.UserID
		windows[i].Provider = connection.Provider
	}
	if err := s.repo.ReplaceBusyWindows(ctx, connection, windows); err != nil {
		return err
	}
	return s.repo.MarkConnectionSynced(ctx, connection.WorkspaceID, connection.ID, s.now().UTC())
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
	path := fmt.Sprintf("%s/%s/settings/workspace/integrations/calendar", base, url.PathEscape(workspaceSlug))
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

func normalizeScheduleBlockInput(input CoreScheduleBlockInput) (CoreScheduleBlockInput, error) {
	if input.WorkspaceID == uuid.Nil || input.UserID == uuid.Nil {
		return CoreScheduleBlockInput{}, ErrInvalidScheduleBlock
	}
	if err := validateScheduleRange(input.StartAt, input.EndAt); err != nil {
		return CoreScheduleBlockInput{}, err
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
