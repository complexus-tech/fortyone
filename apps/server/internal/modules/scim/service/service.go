package scimservice

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"strings"
	"time"

	domain "github.com/complexus-tech/projects-api/internal/modules/scim/domain"
	"github.com/google/uuid"
)

type Repository interface {
	Authenticate(context.Context, string, []byte) (domain.Scope, error)
	AuthorizeAdmin(context.Context, domain.AdminScope) error
	Status(context.Context, domain.AdminScope) (domain.Status, error)
	Mint(context.Context, domain.AdminScope, string, []byte, string, time.Time) (domain.Credential, error)
	Revoke(context.Context, domain.AdminScope, uuid.UUID) error
	List(context.Context, domain.Scope, domain.Page) ([]domain.User, int64, error)
	Get(context.Context, domain.Scope, uuid.UUID) (domain.User, error)
	Create(context.Context, domain.Scope, domain.Mutation) (domain.User, error)
	Update(context.Context, domain.Scope, uuid.UUID, int64, domain.Mutation, bool) (domain.User, error)
	PendingSeats(context.Context, domain.AdminScope) (int64, error)
	FinishSeats(context.Context, domain.AdminScope, int64, bool) error
}
type SeatManager interface {
	UpdateSubscriptionSeats(context.Context, uuid.UUID) error
}
type Service struct {
	repo  Repository
	seats SeatManager
	key   []byte
	now   func() time.Time
}

func New(repo Repository, seats SeatManager, secret string) (*Service, error) {
	if len(secret) < 16 {
		return nil, fmt.Errorf("SCIM token signing secret is too short")
	}
	key := hmac.New(sha256.New, []byte(secret))
	_, _ = key.Write([]byte("fortyone.workspace-scim.credentials.v1"))
	return &Service{repo: repo, seats: seats, key: key.Sum(nil), now: time.Now}, nil
}
func (s *Service) digest(token string) []byte {
	mac := hmac.New(sha256.New, s.key)
	_, _ = mac.Write([]byte(token))
	return mac.Sum(nil)
}
func (s *Service) Authenticate(ctx context.Context, slug, token string) (domain.Scope, error) {
	if !strings.HasPrefix(token, "f41_scim_") || len(token) != 52 {
		return domain.Scope{}, domain.ErrUnauthorized
	}
	if _, err := base64.RawURLEncoding.DecodeString(strings.TrimPrefix(token, "f41_scim_")); err != nil {
		return domain.Scope{}, domain.ErrUnauthorized
	}
	return s.repo.Authenticate(ctx, slug, s.digest(token))
}
func (s *Service) Status(ctx context.Context, scope domain.AdminScope) (domain.Status, error) {
	return s.repo.Status(ctx, scope)
}
func (s *Service) Mint(ctx context.Context, scope domain.AdminScope, in domain.MintInput) (domain.Minted, error) {
	name := strings.TrimSpace(in.Name)
	if len(name) < 1 || len(name) > 100 || in.LifetimeDays < 1 || in.LifetimeDays > 365 {
		return domain.Minted{}, domain.ErrInvalidInput
	}
	random := make([]byte, 32)
	if _, err := rand.Read(random); err != nil {
		return domain.Minted{}, err
	}
	token := "f41_scim_" + base64.RawURLEncoding.EncodeToString(random)
	credential, err := s.repo.Mint(ctx, scope, name, s.digest(token), token[:17], s.now().UTC().Add(time.Duration(in.LifetimeDays)*24*time.Hour))
	if err != nil {
		return domain.Minted{}, err
	}
	return domain.Minted{Credential: credential, Token: token}, nil
}
func (s *Service) Revoke(ctx context.Context, scope domain.AdminScope, id uuid.UUID) error {
	return s.repo.Revoke(ctx, scope, id)
}
func (s *Service) List(ctx context.Context, scope domain.Scope, page domain.Page) ([]domain.User, int64, error) {
	return s.repo.List(ctx, scope, page)
}
func (s *Service) Get(ctx context.Context, scope domain.Scope, id uuid.UUID) (domain.User, error) {
	return s.repo.Get(ctx, scope, id)
}
func (s *Service) Create(ctx context.Context, scope domain.Scope, input domain.Input) (domain.User, error) {
	mutation, err := validate(input, true)
	if err != nil {
		return domain.User{}, err
	}
	user, err := s.repo.Create(ctx, scope, mutation)
	if err == nil {
		s.reconcile(ctx, domain.AdminScope{WorkspaceID: scope.WorkspaceID, ActorID: scope.IssuerID})
	}
	return user, err
}
func (s *Service) Replace(ctx context.Context, scope domain.Scope, id uuid.UUID, input domain.Input) (domain.User, error) {
	old, err := s.repo.Get(ctx, scope, id)
	if err != nil {
		return domain.User{}, err
	}
	mutation, err := validate(input, false)
	if err != nil {
		return domain.User{}, err
	}
	user, err := s.repo.Update(ctx, scope, id, old.Version, mutation, false)
	if err == nil {
		s.reconcile(ctx, domain.AdminScope{WorkspaceID: scope.WorkspaceID, ActorID: scope.IssuerID})
	}
	return user, err
}
func (s *Service) Delete(ctx context.Context, scope domain.Scope, id uuid.UUID) error {
	old, err := s.repo.Get(ctx, scope, id)
	if err != nil {
		return err
	}
	_, err = s.repo.Update(ctx, scope, id, old.Version, domain.Mutation{UserName: old.UserName, ExternalID: old.ExternalID, Profile: old.Profile, Active: false}, true)
	if err == nil {
		s.reconcile(ctx, domain.AdminScope{WorkspaceID: scope.WorkspaceID, ActorID: scope.IssuerID})
	}
	return err
}
func (s *Service) RetrySeats(ctx context.Context, scope domain.AdminScope) error {
	if err := s.repo.AuthorizeAdmin(ctx, scope); err != nil {
		return err
	}
	return s.reconcile(ctx, scope)
}
func (s *Service) reconcile(ctx context.Context, scope domain.AdminScope) error {
	generation, err := s.repo.PendingSeats(ctx, scope)
	if err != nil {
		return err
	}
	if generation == 0 {
		return nil
	}
	if s.seats == nil {
		_ = s.repo.FinishSeats(ctx, scope, generation, false)
		return fmt.Errorf("seat synchronization is unavailable")
	}
	err = s.seats.UpdateSubscriptionSeats(ctx, scope.WorkspaceID)
	finishErr := s.repo.FinishSeats(ctx, scope, generation, err == nil)
	if err != nil {
		return fmt.Errorf("seat synchronization failed; retry from workspace security")
	}
	return finishErr
}
