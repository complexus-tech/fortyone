package googledrive

import (
	"context"
	"testing"
	"time"

	"github.com/complexus-tech/projects-api/internal/modules/googledrive/domain"
	"github.com/complexus-tech/projects-api/internal/platform/credentialvault"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

type providerClientStub struct {
	findResult      *domain.ProviderFile
	getFileResult   domain.ProviderFile
	getFileErr      error
	readFileResult  ProviderContent
	readFileErr     error
	thumbnailResult Preview
	thumbnailErr    error
	exchangeToken   domain.OAuthToken
	exchangeScopes  []string
	exchangeErr     error
	userInfoResult  ProviderUser
	userInfoErr     error
	userInfoResults []ProviderUser
	userInfoErrors  []error
	events          *[]string
	findCalls       int
	createCalls     int
	exchangeCalls   int
	userInfoCalls   int
	revokeCalls     int
	readFileCalls   int
	getFileCalls    int
	thumbnailCalls  int
	thumbnailToken  string
	thumbnailURL    string
	thumbnailLimit  int64
	revokedToken    string
	revokeErr       error
}

type oauthRepositoryStub struct {
	Repository
	state              domain.OAuthState
	activeAccount      domain.Account
	activeAccountFound bool
	activeAccountErr   error
	connection         domain.Connection
	connectionFound    bool
	upsertErr          error
	enqueueErr         error
	enqueuedRevocation domain.Revocation
	events             *[]string
	consumeCalls       int
	upsertCalls        int
	enqueueCalls       int
	disconnectCalls    int
}

type failingSealVault struct {
	CredentialVault
	err error
}

func (vault failingSealVault) Seal(credentialvault.Context, []byte) (string, error) {
	return "", vault.err
}

type referenceRepositoryStub struct {
	Repository
	reference                  domain.FileReference
	targetMutable              bool
	revalidatedGrantGeneration uuid.UUID
	getReferenceWorkspaceID    uuid.UUID
	getReferenceUserID         uuid.UUID
	getReferenceID             uuid.UUID
	revalidatedWorkspaceID     uuid.UUID
	revalidatedUserID          uuid.UUID
	revalidatedAccountID       uuid.UUID
	revalidateCalls            int
	deleteGrantCalls           int
	deletedGrantWorkspaceID    uuid.UUID
	deletedGrantUserID         uuid.UUID
	deletedGrantAccountID      uuid.UUID
	deletedGrantReferenceID    uuid.UUID
	deletedGrantGeneration     uuid.UUID
	markUnavailableCalls       int
	unavailableWorkspaceID     uuid.UUID
	unavailableReferenceID     uuid.UUID
	failedImportCalls          int
}

func (repo *referenceRepositoryStub) GetReference(
	_ context.Context,
	workspaceID uuid.UUID,
	userID uuid.UUID,
	referenceID uuid.UUID,
) (domain.FileReference, error) {
	repo.getReferenceWorkspaceID = workspaceID
	repo.getReferenceUserID = userID
	repo.getReferenceID = referenceID
	return repo.reference, nil
}

func (repo *referenceRepositoryStub) TargetMutable(
	context.Context,
	uuid.UUID,
	uuid.UUID,
	domain.TargetType,
	uuid.UUID,
) (bool, error) {
	return repo.targetMutable, nil
}

func (repo *referenceRepositoryStub) RevalidateReference(
	_ context.Context,
	workspaceID uuid.UUID,
	userID uuid.UUID,
	accountID uuid.UUID,
	_ domain.FileReference,
	_ domain.ProviderFile,
) (uuid.UUID, error) {
	repo.revalidateCalls++
	repo.revalidatedWorkspaceID = workspaceID
	repo.revalidatedUserID = userID
	repo.revalidatedAccountID = accountID
	return repo.revalidatedGrantGeneration, nil
}

func (repo *referenceRepositoryStub) DeleteReferenceGrant(
	_ context.Context,
	workspaceID, userID, accountID, referenceID, grantGeneration uuid.UUID,
) error {
	repo.deleteGrantCalls++
	repo.deletedGrantWorkspaceID = workspaceID
	repo.deletedGrantUserID = userID
	repo.deletedGrantAccountID = accountID
	repo.deletedGrantReferenceID = referenceID
	repo.deletedGrantGeneration = grantGeneration
	return nil
}

func (repo *referenceRepositoryStub) MarkReferenceUnavailable(
	_ context.Context,
	workspaceID, referenceID uuid.UUID,
) error {
	repo.markUnavailableCalls++
	repo.unavailableWorkspaceID = workspaceID
	repo.unavailableReferenceID = referenceID
	return nil
}

func (repo *referenceRepositoryStub) CreateImportOperation(
	_ context.Context,
	operation domain.ImportOperation,
) (domain.ImportOperation, bool, error) {
	now := time.Now().UTC()
	operation.ID = uuid.New()
	operation.Status = domain.ImportOperationPending
	operation.CreatedAt = now
	operation.UpdatedAt = now
	return operation, true, nil
}

func (repo *referenceRepositoryStub) FailImportOperation(
	context.Context,
	uuid.UUID,
	uuid.UUID,
	string,
) error {
	repo.failedImportCalls++
	return nil
}

func (repo *oauthRepositoryStub) ConsumeOAuthState(context.Context, string, time.Time) (domain.OAuthState, error) {
	repo.consumeCalls++
	return repo.state, nil
}

func (repo *oauthRepositoryStub) WithinProviderUserLifecycle(
	ctx context.Context,
	_ uuid.UUID,
	operation func(context.Context) error,
) error {
	repo.record("user_lock_begin")
	err := operation(ctx)
	repo.record("user_lock_end")
	return err
}

func (repo *oauthRepositoryStub) WithinProviderSubjectLifecycle(
	ctx context.Context,
	_ string,
	operation func(context.Context) error,
) error {
	repo.record("subject_lock_begin")
	err := operation(ctx)
	repo.record("subject_lock_end")
	return err
}

func (repo *oauthRepositoryStub) UpsertConnection(context.Context, uuid.UUID, domain.Account) (domain.Connection, error) {
	repo.upsertCalls++
	repo.record("upsert")
	return domain.Connection{}, repo.upsertErr
}

func (repo *oauthRepositoryStub) EnqueueRevocation(
	_ context.Context,
	revocation domain.Revocation,
) (domain.RevocationCandidate, error) {
	repo.enqueueCalls++
	repo.enqueuedRevocation = revocation
	if repo.enqueueErr != nil {
		return domain.RevocationCandidate{}, repo.enqueueErr
	}
	return domain.RevocationCandidate{
		ID: uuid.New(), UserID: revocation.UserID, GoogleSubject: revocation.GoogleSubject,
	}, nil
}

func (repo *oauthRepositoryStub) GetActiveAccountBySubject(context.Context, string) (domain.Account, error) {
	repo.record("get_subject")
	if repo.activeAccountErr != nil {
		return domain.Account{}, repo.activeAccountErr
	}
	if !repo.activeAccountFound {
		return domain.Account{}, domain.ErrNotFound
	}
	return repo.activeAccount, nil
}

func (repo *oauthRepositoryStub) GetConnection(context.Context, uuid.UUID, uuid.UUID) (domain.Connection, error) {
	repo.record("get_connection")
	if !repo.connectionFound {
		return domain.Connection{}, domain.ErrNotFound
	}
	return repo.connection, nil
}

func (repo *oauthRepositoryStub) Disconnect(context.Context, uuid.UUID, uuid.UUID) (bool, error) {
	repo.disconnectCalls++
	repo.record("disconnect")
	return true, nil
}

func (repo *oauthRepositoryStub) record(event string) {
	if repo.events != nil {
		*repo.events = append(*repo.events, event)
	}
}

func configuredOAuthTestConfig(t *testing.T) Config {
	t.Helper()
	vault, err := credentialvault.NewFromEncodedKeyring(
		credentialvault.DevelopmentKeyID,
		credentialvault.DevelopmentKeyVersion,
		credentialvault.DevelopmentEncodedKeys,
	)
	require.NoError(t, err)
	return Config{
		ClientID: "client-id", ClientSecret: "client-secret",
		RedirectURL:  "https://api.fortyone.app/integrations/google-drive/callback",
		PickerAPIKey: "picker-key", AppID: "123456789", WebsiteURL: "https://fortyone.app",
		Credentials: vault,
	}
}

func pointer(value string) *string { return &value }

func pointerUUID(value uuid.UUID) *uuid.UUID { return &value }

func (client *providerClientStub) AuthorizationURL(string, string) (string, error) {
	return "", nil
}
func (client *providerClientStub) Exchange(context.Context, string, string) (domain.OAuthToken, []string, error) {
	client.exchangeCalls++
	client.record("exchange")
	return client.exchangeToken, client.exchangeScopes, client.exchangeErr
}
func (client *providerClientStub) Refresh(context.Context, string) (domain.OAuthToken, error) {
	return domain.OAuthToken{}, nil
}
func (client *providerClientStub) UserInfo(context.Context, string) (ProviderUser, error) {
	callIndex := client.userInfoCalls
	client.userInfoCalls++
	client.record("user_info")
	if callIndex < len(client.userInfoErrors) && client.userInfoErrors[callIndex] != nil {
		return ProviderUser{}, client.userInfoErrors[callIndex]
	}
	if callIndex < len(client.userInfoResults) {
		return client.userInfoResults[callIndex], nil
	}
	return client.userInfoResult, client.userInfoErr
}
func (client *providerClientStub) Revoke(_ context.Context, token string) error {
	client.revokeCalls++
	client.revokedToken = token
	return client.revokeErr
}
func (client *providerClientStub) GetFile(context.Context, string, string, *string) (domain.ProviderFile, error) {
	client.getFileCalls++
	return client.getFileResult, client.getFileErr
}
func (client *providerClientStub) CreateFile(context.Context, string, domain.FileType, string, string) (domain.ProviderFile, error) {
	client.createCalls++
	return domain.ProviderFile{}, nil
}
func (client *providerClientStub) FindCreatedFile(context.Context, string, string) (*domain.ProviderFile, error) {
	client.findCalls++
	return client.findResult, nil
}
func (client *providerClientStub) PopulateFile(context.Context, string, domain.ProviderFile, string) error {
	return nil
}
func (client *providerClientStub) ReadFile(context.Context, string, domain.ProviderFile, int64) (ProviderContent, error) {
	client.readFileCalls++
	return client.readFileResult, client.readFileErr
}
func (client *providerClientStub) ReadThumbnail(_ context.Context, token, thumbnailURL string, limit int64) (Preview, error) {
	client.thumbnailCalls++
	client.thumbnailToken = token
	client.thumbnailURL = thumbnailURL
	client.thumbnailLimit = limit
	return client.thumbnailResult, client.thumbnailErr
}

func (client *providerClientStub) record(event string) {
	if client.events != nil {
		*client.events = append(*client.events, event)
	}
}
