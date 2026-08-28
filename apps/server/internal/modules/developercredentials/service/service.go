package developercredentials

import (
	"context"
	"errors"
	"time"

	developercredentialsdomain "github.com/complexus-tech/projects-api/internal/modules/developercredentials/domain"
	"github.com/google/uuid"
)

type Repository interface {
	EnsureHumanPrincipal(context.Context, developercredentialsdomain.EnsureHumanPrincipal) (uuid.UUID, error)
	ResolveHumanPrincipal(context.Context, uuid.UUID, uuid.UUID) (uuid.UUID, error)
	CreatePersonalAccessToken(context.Context, developercredentialsdomain.CreatePersonalToken) (developercredentialsdomain.Credential, error)
	ListPersonalAccessTokens(context.Context, uuid.UUID, uuid.UUID) ([]developercredentialsdomain.Credential, error)
	RotatePersonalAccessToken(context.Context, developercredentialsdomain.RotateCredential) (developercredentialsdomain.Credential, error)
	RevokePersonalAccessToken(context.Context, developercredentialsdomain.RevokeCredential) error

	CreateServiceAccount(context.Context, developercredentialsdomain.CreateServiceAccount) (developercredentialsdomain.ServiceAccount, error)
	ListServiceAccounts(context.Context, uuid.UUID, uuid.UUID) ([]developercredentialsdomain.ServiceAccount, error)
	DisableServiceAccount(context.Context, developercredentialsdomain.DisableServiceAccount) error
	CreateServiceAccountKey(context.Context, developercredentialsdomain.CreateServiceAccountKey) (developercredentialsdomain.Credential, error)
	ListServiceAccountKeys(context.Context, uuid.UUID, uuid.UUID, uuid.UUID) ([]developercredentialsdomain.Credential, error)
	RotateServiceAccountKey(context.Context, developercredentialsdomain.RotateCredential) (developercredentialsdomain.Credential, error)
	RevokeServiceAccountKey(context.Context, developercredentialsdomain.RevokeCredential) error

	LookupCredential(context.Context, string, developercredentialsdomain.CredentialKind, int16, time.Time) (developercredentialsdomain.VerificationRecord, error)
	ConfirmCredentialActiveAndTouch(context.Context, uuid.UUID, time.Time, time.Time) error
}

type Clock interface {
	Now() time.Time
}

type IDGenerator interface {
	NewID() (uuid.UUID, error)
}

type WallClock struct{}

func (WallClock) Now() time.Time {
	return time.Now().UTC()
}

type RandomIDGenerator struct{}

func (RandomIDGenerator) NewID() (uuid.UUID, error) {
	return uuid.NewRandom()
}

type Service struct {
	repository Repository
	tokens     *TokenManager
	clock      Clock
	ids        IDGenerator
}

func New(repository Repository, tokens *TokenManager, clock Clock, ids IDGenerator) (*Service, error) {
	if repository == nil {
		return nil, errors.New("developer credential repository is required")
	}
	if tokens == nil {
		return nil, errors.New("developer credential token manager is required")
	}
	if clock == nil {
		return nil, errors.New("developer credential clock is required")
	}
	if ids == nil {
		return nil, errors.New("developer credential ID generator is required")
	}
	return &Service{repository: repository, tokens: tokens, clock: clock, ids: ids}, nil
}

func (service *Service) nextID() (uuid.UUID, error) {
	id, err := service.ids.NewID()
	if err != nil {
		return uuid.Nil, err
	}
	if id == uuid.Nil {
		return uuid.Nil, errors.New("developer credential ID generator returned a zero UUID")
	}
	return id, nil
}
