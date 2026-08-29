package developercredentials

import (
	"context"
	"errors"
	"time"

	developercredentialsdomain "github.com/complexus-tech/projects-api/internal/modules/developercredentials/domain"
	"github.com/google/uuid"
)

var errUnexpectedRepositoryCall = errors.New("unexpected developer credential repository call")

type fakeRepository struct {
	ensureHumanPrincipal    func(developercredentialsdomain.EnsureHumanPrincipal) (uuid.UUID, error)
	resolveHumanPrincipal   func(uuid.UUID, uuid.UUID) (uuid.UUID, error)
	createServiceAccount    func(developercredentialsdomain.CreateServiceAccount) (developercredentialsdomain.ServiceAccount, error)
	createServiceAccountKey func(developercredentialsdomain.CreateServiceAccountKey) (developercredentialsdomain.Credential, error)
	lookupCredential        func(string, developercredentialsdomain.CredentialKind, int16, time.Time) (developercredentialsdomain.VerificationRecord, error)
	confirmCredential       func(uuid.UUID, time.Time, time.Time) error
}

func (fake *fakeRepository) EnsureHumanPrincipal(_ context.Context, command developercredentialsdomain.EnsureHumanPrincipal) (uuid.UUID, error) {
	if fake.ensureHumanPrincipal == nil {
		return uuid.Nil, errUnexpectedRepositoryCall
	}
	return fake.ensureHumanPrincipal(command)
}
func (fake *fakeRepository) ResolveHumanPrincipal(_ context.Context, workspaceID uuid.UUID, userID uuid.UUID) (uuid.UUID, error) {
	if fake.resolveHumanPrincipal == nil {
		return uuid.Nil, errUnexpectedRepositoryCall
	}
	return fake.resolveHumanPrincipal(workspaceID, userID)
}

func (fake *fakeRepository) CreatePersonalAccessToken(context.Context, developercredentialsdomain.CreatePersonalToken) (developercredentialsdomain.Credential, error) {
	return developercredentialsdomain.Credential{}, errUnexpectedRepositoryCall
}
func (fake *fakeRepository) ListPersonalAccessTokens(context.Context, uuid.UUID, uuid.UUID) ([]developercredentialsdomain.Credential, error) {
	return nil, errUnexpectedRepositoryCall
}
func (fake *fakeRepository) RotatePersonalAccessToken(context.Context, developercredentialsdomain.RotateCredential) (developercredentialsdomain.Credential, error) {
	return developercredentialsdomain.Credential{}, errUnexpectedRepositoryCall
}
func (fake *fakeRepository) RevokePersonalAccessToken(context.Context, developercredentialsdomain.RevokeCredential) error {
	return errUnexpectedRepositoryCall
}
func (fake *fakeRepository) CreateServiceAccount(_ context.Context, command developercredentialsdomain.CreateServiceAccount) (developercredentialsdomain.ServiceAccount, error) {
	if fake.createServiceAccount == nil {
		return developercredentialsdomain.ServiceAccount{}, errUnexpectedRepositoryCall
	}
	return fake.createServiceAccount(command)
}
func (fake *fakeRepository) ListServiceAccounts(context.Context, uuid.UUID, uuid.UUID) ([]developercredentialsdomain.ServiceAccount, error) {
	return nil, errUnexpectedRepositoryCall
}
func (fake *fakeRepository) DisableServiceAccount(context.Context, developercredentialsdomain.DisableServiceAccount) error {
	return errUnexpectedRepositoryCall
}
func (fake *fakeRepository) CreateServiceAccountKey(_ context.Context, command developercredentialsdomain.CreateServiceAccountKey) (developercredentialsdomain.Credential, error) {
	if fake.createServiceAccountKey == nil {
		return developercredentialsdomain.Credential{}, errUnexpectedRepositoryCall
	}
	return fake.createServiceAccountKey(command)
}
func (fake *fakeRepository) ListServiceAccountKeys(context.Context, uuid.UUID, uuid.UUID, uuid.UUID) ([]developercredentialsdomain.Credential, error) {
	return nil, errUnexpectedRepositoryCall
}
func (fake *fakeRepository) RotateServiceAccountKey(context.Context, developercredentialsdomain.RotateCredential) (developercredentialsdomain.Credential, error) {
	return developercredentialsdomain.Credential{}, errUnexpectedRepositoryCall
}
func (fake *fakeRepository) RevokeServiceAccountKey(context.Context, developercredentialsdomain.RevokeCredential) error {
	return errUnexpectedRepositoryCall
}
func (fake *fakeRepository) LookupCredential(_ context.Context, prefix string, kind developercredentialsdomain.CredentialKind, version int16, at time.Time) (developercredentialsdomain.VerificationRecord, error) {
	if fake.lookupCredential == nil {
		return developercredentialsdomain.VerificationRecord{}, errUnexpectedRepositoryCall
	}
	return fake.lookupCredential(prefix, kind, version, at)
}
func (fake *fakeRepository) ConfirmCredentialActiveAndTouch(_ context.Context, credentialID uuid.UUID, usedAt time.Time, touchBefore time.Time) error {
	if fake.confirmCredential == nil {
		return errUnexpectedRepositoryCall
	}
	return fake.confirmCredential(credentialID, usedAt, touchBefore)
}

type fixedClock struct{ now time.Time }

func (clock fixedClock) Now() time.Time { return clock.now }

type sequenceIDGenerator struct {
	values []uuid.UUID
	index  int
}

func (generator *sequenceIDGenerator) NewID() (uuid.UUID, error) {
	if generator.index >= len(generator.values) {
		return uuid.Nil, errors.New("test ID sequence exhausted")
	}
	value := generator.values[generator.index]
	generator.index++
	return value, nil
}

func testTokenManager(t testFataler) *TokenManager {
	t.Helper()
	manager, err := NewTokenManager(TokenKeyringConfig{
		Active: developercredentialsdomain.DigestKeyRef{ID: "test", Version: 1},
		Keys: []DigestKey{{
			Ref:      developercredentialsdomain.DigestKeyRef{ID: "test", Version: 1},
			Material: []byte("0123456789abcdef0123456789abcdef"),
		}},
	})
	if err != nil {
		t.Fatalf("create test token manager: %v", err)
	}
	return manager
}

type testFataler interface {
	Helper()
	Fatalf(string, ...any)
}
