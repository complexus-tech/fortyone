package googledrive

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/complexus-tech/projects-api/internal/modules/googledrive/domain"
	"github.com/complexus-tech/projects-api/internal/platform/credentialvault"
	"github.com/complexus-tech/projects-api/internal/platform/workspaceurl"
	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/google/uuid"
)

const (
	oauthStateTTL                = 10 * time.Minute
	googleDriveFileScope         = "https://www.googleapis.com/auth/drive.file"
	tokenRefreshSkew             = 5 * time.Minute
	maxAttachedFiles             = 20
	maxProviderFileIDBytes       = 512
	maxResourceKeyBytes          = 512
	maxTitleRunes                = 255
	maxContentBytes        int64 = 1 << 20
	maxPreviewBytes        int64 = 5 << 20
	createOperationLease         = 2 * time.Minute
)

var oauthScopes = []string{
	"openid",
	"email",
	"profile",
	googleDriveFileScope,
}

type Service struct {
	log    *logger.Logger
	repo   Repository
	config Config
	client ProviderClient
	now    func() time.Time
}

func New(log *logger.Logger, repo Repository, config Config) *Service {
	httpClient := &http.Client{Timeout: 25 * time.Second}
	return &Service{
		log: log, repo: repo, config: config,
		client: newGoogleClient(httpClient, config, oauthScopes), now: time.Now,
	}
}

func (service *Service) configured() bool {
	return service != nil && service.repo != nil && service.config.Credentials != nil &&
		strings.TrimSpace(service.config.ClientID) != "" &&
		strings.TrimSpace(service.config.ClientSecret) != "" &&
		strings.TrimSpace(service.config.RedirectURL) != "" &&
		strings.TrimSpace(service.config.PickerAPIKey) != "" &&
		strings.TrimSpace(service.config.AppID) != ""
}

func (service *Service) GetIntegration(ctx context.Context, workspaceID, userID uuid.UUID) (Integration, error) {
	result := Integration{Configured: service.configured(), Status: "disconnected"}
	connection, err := service.repo.GetConnection(ctx, workspaceID, userID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return result, nil
		}
		return Integration{}, err
	}
	result.Connected = true
	result.Email = &connection.Email
	result.RequiresReauthorization = connection.RequiresReauthorization
	result.Status = "connected"
	if connection.RequiresReauthorization {
		result.Status = "reauthorization_required"
	}
	return result, nil
}

func (service *Service) CreateConnectSession(
	ctx context.Context,
	workspaceID, userID uuid.UUID,
	workspaceSlug, requestedReturnURL string,
) (string, error) {
	if !service.configured() {
		return "", domain.ErrNotConfigured
	}
	returnURL := service.validatedReturnURL(workspaceSlug, requestedReturnURL)
	state, verifier, err := service.createOAuthState(ctx, workspaceID, userID, workspaceSlug, returnURL)
	if err != nil {
		return "", err
	}
	return service.client.AuthorizationURL(state, verifier)
}

func (service *Service) CompleteOAuth(
	ctx context.Context,
	actorID uuid.UUID,
	code, rawState, providerError string,
) (string, error) {
	stored, err := service.consumeOAuthState(ctx, rawState)
	if err != nil {
		return "", fmt.Errorf("the Google Drive connection session expired; try again: %w", err)
	}
	failureCode := "connection_failed"
	if providerError == "access_denied" {
		failureCode = "access_denied"
	}
	failureURL := appendQuery(service.oauthReturnURL(stored), "google_drive_error", failureCode)
	if actorID == uuid.Nil || stored.UserID == uuid.Nil || actorID != stored.UserID {
		return failureURL, domain.ErrForbidden
	}
	if providerError != "" {
		return failureURL, errors.New("Google Drive authorization was not completed")
	}
	if strings.TrimSpace(code) == "" {
		return failureURL, errors.New("Google Drive OAuth callback is missing its authorization code")
	}
	if !service.configured() {
		return failureURL, domain.ErrNotConfigured
	}
	var successURL string
	err = service.repo.WithinProviderUserLifecycle(ctx, stored.UserID, func(lifecycleCtx context.Context) error {
		if err := service.completeOAuthWithinUserLifecycle(lifecycleCtx, stored, code); err != nil {
			return err
		}
		successURL = appendQuery(service.oauthReturnURL(stored), "google_drive_connected", "1")
		return nil
	})
	if err != nil {
		return failureURL, err
	}
	return successURL, nil
}

// completeOAuthWithinUserLifecycle runs while the repository owns a
// session-scoped user provider gate. Provider calls are deliberately outside
// database transactions, while disconnect and reconnect still have one total
// order across processes.
func (service *Service) completeOAuthWithinUserLifecycle(
	ctx context.Context,
	stored domain.OAuthState,
	code string,
) error {
	token, scopes, err := service.client.Exchange(ctx, code, stored.CodeVerifier)
	if err != nil {
		service.logProviderError(ctx, "exchange Google Drive OAuth code", err, stored.WorkspaceID, stored.UserID)
		return err
	}
	providerUser, err := service.client.UserInfo(ctx, token.AccessToken)
	if err != nil {
		service.logProviderError(ctx, "load Google account identity", err, stored.WorkspaceID, stored.UserID)
		// Without a verified subject, revocation could invalidate a grant owned
		// by another FortyOne user in the same Google Cloud project.
		return err
	}
	return service.repo.WithinProviderSubjectLifecycle(ctx, providerUser.Subject, func(subjectCtx context.Context) error {
		return service.completeOAuthWithinSubjectLifecycle(subjectCtx, stored, token, scopes, providerUser)
	})
}

// completeOAuthWithinSubjectLifecycle runs while both the local-user and
// global Google-subject provider gates are held. A second identity read detects
// an exchange that raced a just-finished cross-user revocation before any
// credential is persisted.
func (service *Service) completeOAuthWithinSubjectLifecycle(
	ctx context.Context,
	stored domain.OAuthState,
	token domain.OAuthToken,
	scopes []string,
	providerUser ProviderUser,
) error {
	verifiedUser, err := service.client.UserInfo(ctx, token.AccessToken)
	if err != nil {
		service.logProviderError(ctx, "revalidate Google account identity", err, stored.WorkspaceID, stored.UserID)
		_, ownershipErr := service.repo.GetActiveAccountBySubject(ctx, providerUser.Subject)
		switch {
		case ownershipErr == nil:
			// The project-wide grant is already represented by an active local
			// owner. Revoking this token could disconnect that owner.
			return err
		case errors.Is(ownershipErr, domain.ErrNotFound):
			return errors.Join(
				err,
				service.stageUnpersistedGrantRevocation(ctx, stored, providerUser, token, true),
			)
		default:
			// Fail closed when ownership cannot be established. A dropped
			// cleanup is safer than revoking another user's active grant.
			return errors.Join(err, fmt.Errorf("verify Google account ownership before cleanup: %w", ownershipErr))
		}
	}
	if verifiedUser.Subject != providerUser.Subject {
		return errors.New("Google account identity changed during connection; try again")
	}
	providerUser = verifiedUser

	existingAccount, accountErr := service.repo.GetActiveAccountBySubject(ctx, providerUser.Subject)
	if accountErr != nil && !errors.Is(accountErr, domain.ErrNotFound) {
		return accountErr
	}
	hasMatchingAccount := accountErr == nil && existingAccount.UserID == stored.UserID
	if accountErr == nil && existingAccount.UserID != stored.UserID {
		// Revoking this just-exchanged token would also revoke the existing
		// owner's project-wide Google grant.
		return domain.ErrAccountOwned
	}
	revokeUnpersistedGrant := func(cause error) error {
		if hasMatchingAccount {
			return cause
		}
		return errors.Join(cause, service.stageUnpersistedGrantRevocation(ctx, stored, providerUser, token, true))
	}
	if !hasOAuthScope(scopes, googleDriveFileScope) {
		return revokeUnpersistedGrant(errors.New("Google Drive authorization did not grant the required file access; reconnect and approve access"))
	}
	existingConnection, connectionErr := service.repo.GetConnection(ctx, stored.WorkspaceID, stored.UserID)
	if connectionErr != nil && !errors.Is(connectionErr, domain.ErrNotFound) {
		return revokeUnpersistedGrant(connectionErr)
	}
	if connectionErr == nil && existingConnection.GoogleSubject != providerUser.Subject {
		return revokeUnpersistedGrant(errors.New("disconnect the current Google Drive account before connecting a different account"))
	}
	if token.RefreshToken == "" {
		if hasMatchingAccount {
			if previousToken, openErr := service.openToken(existingAccount); openErr == nil {
				token.RefreshToken = previousToken.RefreshToken
			}
		}
	}
	if token.RefreshToken == "" {
		return revokeUnpersistedGrant(errors.New("Google did not issue an offline refresh token; reconnect and approve access"))
	}
	generation := uuid.New()
	account := domain.Account{
		UserID: stored.UserID, GoogleSubject: providerUser.Subject,
		Email: providerUser.Email, DisplayName: providerUser.DisplayName,
		InstallationGeneration: generation, Scopes: scopes, ExpiresAt: token.Expiry,
		CredentialVersion: int16(credentialvault.CurrentVersion),
	}
	account.CredentialPayload, err = service.sealToken(account, token)
	if err != nil {
		return revokeUnpersistedGrant(err)
	}
	if _, err := service.repo.UpsertConnection(ctx, stored.WorkspaceID, account); err != nil {
		service.logProviderError(ctx, "store Google Drive connection", err, stored.WorkspaceID, stored.UserID)
		if errors.Is(err, domain.ErrAccountOwned) || hasMatchingAccount {
			return err
		}
		// A transaction commit error can be ambiguous: the connection may be
		// durable even though the client observed an error. The outbox enqueue
		// performs its own ownership check, but an enqueue failure must never
		// fall back to project-wide revocation based only on the stale pre-write
		// absence check.
		return errors.Join(
			err,
			service.stageUnpersistedGrantRevocation(ctx, stored, providerUser, token, false),
		)
	}
	return nil
}

func (service *Service) PickerSession(ctx context.Context, workspaceID, userID uuid.UUID, workspaceSlug, origin string) (PickerSession, error) {
	if !service.configured() {
		return PickerSession{}, domain.ErrNotConfigured
	}
	validatedOrigin := service.validatedWorkspaceOrigin(workspaceSlug, origin)
	if strings.TrimSpace(origin) != "" && validatedOrigin == "" {
		return PickerSession{}, domain.ErrForbidden
	}
	if validatedOrigin == "" {
		validatedOrigin = service.workspaceOrigin(workspaceSlug)
	}
	if validatedOrigin == "" {
		return PickerSession{}, domain.ErrNotConfigured
	}
	_, token, err := service.connectionToken(ctx, workspaceID, userID)
	if err != nil {
		return PickerSession{}, err
	}
	result := PickerSession{AccessToken: token.AccessToken, APIKey: service.config.PickerAPIKey, AppID: service.config.AppID}
	result.Origin = &validatedOrigin
	return result, nil
}

func (service *Service) AttachFiles(ctx context.Context, input AttachInput) ([]domain.FileReference, error) {
	if !service.configured() {
		return nil, domain.ErrNotConfigured
	}
	if err := validateTarget(input.WorkspaceID, input.UserID, input.TargetType, input.TargetID); err != nil {
		return nil, err
	}
	if len(input.Files) == 0 || len(input.Files) > maxAttachedFiles {
		return nil, domain.ErrInvalidInput
	}
	mutable, err := service.repo.TargetMutable(ctx, input.WorkspaceID, input.UserID, input.TargetType, input.TargetID)
	if err != nil {
		return nil, err
	}
	if !mutable {
		return nil, domain.ErrForbidden
	}
	connection, token, err := service.connectionToken(ctx, input.WorkspaceID, input.UserID)
	if err != nil {
		return nil, err
	}
	providerFiles := make([]domain.ProviderFile, 0, len(input.Files))
	seen := make(map[string]struct{}, len(input.Files))
	for _, selected := range input.Files {
		fileID := strings.TrimSpace(selected.ID)
		if !validProviderFileID(fileID) {
			return nil, domain.ErrInvalidInput
		}
		resourceKey, err := normalizeResourceKey(selected.ResourceKey)
		if err != nil {
			return nil, err
		}
		if _, exists := seen[fileID]; exists {
			continue
		}
		seen[fileID] = struct{}{}
		providerFile, err := service.client.GetFile(ctx, token.AccessToken, fileID, resourceKey)
		if err != nil {
			return nil, service.handleProviderAuthorizationError(ctx, connection, err)
		}
		if providerFile.Trashed || !supportedFile(providerFile.MimeType) {
			return nil, domain.ErrInvalidInput
		}
		providerFiles = append(providerFiles, providerFile)
	}
	return service.repo.AttachFiles(
		ctx, input.WorkspaceID, input.UserID, connection.ID,
		input.TargetType, input.TargetID, providerFiles,
	)
}

func (service *Service) ListFiles(ctx context.Context, workspaceID, userID uuid.UUID, targetType domain.TargetType, targetID uuid.UUID) ([]domain.FileReference, error) {
	if err := validateTarget(workspaceID, userID, targetType, targetID); err != nil {
		return nil, err
	}
	return service.repo.ListReferences(ctx, workspaceID, userID, targetType, targetID)
}

func (service *Service) DeleteFile(ctx context.Context, workspaceID, userID, referenceID uuid.UUID) error {
	if workspaceID == uuid.Nil || userID == uuid.Nil || referenceID == uuid.Nil {
		return domain.ErrInvalidInput
	}
	return service.repo.DeleteReference(ctx, workspaceID, userID, referenceID)
}

func (service *Service) RefreshFile(ctx context.Context, workspaceID, userID, referenceID uuid.UUID) (domain.FileReference, error) {
	if !service.configured() {
		return domain.FileReference{}, domain.ErrNotConfigured
	}
	reference, err := service.repo.GetReference(ctx, workspaceID, userID, referenceID)
	if err != nil {
		return domain.FileReference{}, err
	}
	mutable, err := service.repo.TargetMutable(ctx, workspaceID, userID, reference.TargetType, reference.TargetID)
	if err != nil {
		return domain.FileReference{}, err
	}
	if !mutable {
		return domain.FileReference{}, domain.ErrForbidden
	}
	if reference.Account == nil {
		return domain.FileReference{}, domain.ErrForbidden
	}
	account, token, err := service.accountToken(ctx, workspaceID, userID, *reference.Account)
	if err != nil {
		return domain.FileReference{}, err
	}
	providerFile, err := service.client.GetFile(ctx, token.AccessToken, reference.FileID, reference.ResourceKey)
	if err != nil {
		return domain.FileReference{}, service.handleReferenceProviderError(
			ctx, workspaceID, userID, reference, account, err,
		)
	}
	if providerFile.Trashed {
		return domain.FileReference{}, service.markReferenceUnavailable(ctx, workspaceID, userID, reference.ID)
	}
	if !supportedFile(providerFile.MimeType) {
		return domain.FileReference{}, domain.ErrInvalidInput
	}
	return service.repo.AttachFile(ctx, workspaceID, userID, account.ID, reference.TargetType, reference.TargetID, providerFile)
}

func (service *Service) ReadContent(ctx context.Context, workspaceID, userID, referenceID uuid.UUID) (domain.Content, error) {
	if !service.configured() {
		return domain.Content{}, domain.ErrNotConfigured
	}
	reference, err := service.repo.GetReference(ctx, workspaceID, userID, referenceID)
	if err != nil {
		return domain.Content{}, err
	}
	if reference.Account == nil {
		return domain.Content{}, domain.ErrForbidden
	}
	_, token, err := service.accountToken(ctx, workspaceID, userID, *reference.Account)
	if err != nil {
		return domain.Content{}, err
	}
	providerFile, err := service.client.GetFile(ctx, token.AccessToken, reference.FileID, reference.ResourceKey)
	if err != nil {
		return domain.Content{}, service.handleReferenceProviderError(
			ctx, workspaceID, userID, reference, *reference.Account, err,
		)
	}
	if providerFile.Trashed {
		return domain.Content{}, service.markReferenceUnavailable(ctx, workspaceID, userID, reference.ID)
	}
	if !readableFile(providerFile.MimeType) {
		if _, err := service.repo.RevalidateReference(ctx, workspaceID, userID, reference.Account.ID, reference, providerFile); err != nil {
			return domain.Content{}, err
		}
		return domain.Content{}, domain.ErrInvalidInput
	}
	grantGeneration, err := service.repo.RevalidateReference(ctx, workspaceID, userID, reference.Account.ID, reference, providerFile)
	if err != nil {
		return domain.Content{}, err
	}
	reference.GrantGeneration = &grantGeneration
	content, err := service.client.ReadFile(ctx, token.AccessToken, providerFile, maxContentBytes)
	if err != nil {
		return domain.Content{}, service.handleReferenceProviderError(
			ctx, workspaceID, userID, reference, *reference.Account, err,
		)
	}
	return domain.Content{
		ReferenceID: reference.ID, Name: providerFile.Name, MimeType: providerFile.MimeType,
		WebViewLink: providerFile.WebViewLink, ModifiedAt: providerFile.ModifiedAt,
		Text: content.Text, ContentType: content.ContentType,
		Truncated: content.Truncated, BytesRead: content.BytesRead,
	}, nil
}

func (service *Service) ReadPreview(ctx context.Context, workspaceID, userID, referenceID uuid.UUID) (Preview, error) {
	if !service.configured() {
		return Preview{}, domain.ErrNotConfigured
	}
	reference, err := service.repo.GetReference(ctx, workspaceID, userID, referenceID)
	if err != nil {
		return Preview{}, err
	}
	if reference.Account == nil {
		return Preview{}, domain.ErrForbidden
	}
	account, token, err := service.accountToken(ctx, workspaceID, userID, *reference.Account)
	if err != nil {
		return Preview{}, err
	}
	providerFile, err := service.client.GetFile(ctx, token.AccessToken, reference.FileID, reference.ResourceKey)
	if err != nil {
		return Preview{}, service.handleReferenceProviderError(
			ctx, workspaceID, userID, reference, account, err,
		)
	}
	if providerFile.Trashed {
		return Preview{}, service.markReferenceUnavailable(ctx, workspaceID, userID, reference.ID)
	}
	if !supportedFile(providerFile.MimeType) {
		if _, err := service.repo.RevalidateReference(ctx, workspaceID, userID, account.ID, reference, providerFile); err != nil {
			return Preview{}, err
		}
		return Preview{}, domain.ErrInvalidInput
	}
	grantGeneration, err := service.repo.RevalidateReference(ctx, workspaceID, userID, account.ID, reference, providerFile)
	if err != nil {
		return Preview{}, err
	}
	reference.GrantGeneration = &grantGeneration
	if strings.TrimSpace(providerFile.ThumbnailLink) == "" {
		return Preview{}, domain.ErrNotFound
	}
	preview, err := service.client.ReadThumbnail(ctx, token.AccessToken, providerFile.ThumbnailLink, maxPreviewBytes)
	if err != nil {
		return Preview{}, service.handleReferenceProviderError(
			ctx, workspaceID, userID, reference, account, err,
		)
	}
	return preview, nil
}

func (service *Service) CreateFile(ctx context.Context, input CreateFileInput) (domain.FileReference, error) {
	if !service.configured() {
		return domain.FileReference{}, domain.ErrNotConfigured
	}
	input.Title = strings.TrimSpace(input.Title)
	if err := validateTarget(input.WorkspaceID, input.UserID, input.TargetType, input.TargetID); err != nil ||
		!input.FileType.Valid() || input.Title == "" || utf8.RuneCountInString(input.Title) > maxTitleRunes {
		return domain.FileReference{}, domain.ErrInvalidInput
	}
	if input.IdempotencyKey = strings.TrimSpace(input.IdempotencyKey); input.IdempotencyKey == "" || len(input.IdempotencyKey) > 200 {
		return domain.FileReference{}, domain.ErrInvalidInput
	}
	accessible, err := service.repo.TargetMutable(ctx, input.WorkspaceID, input.UserID, input.TargetType, input.TargetID)
	if err != nil {
		return domain.FileReference{}, err
	}
	if !accessible {
		return domain.FileReference{}, domain.ErrForbidden
	}
	requestHash := createRequestHash(input)
	operation, created, err := service.repo.CreateOperation(ctx, domain.CreateOperation{
		WorkspaceID: input.WorkspaceID, UserID: input.UserID,
		IdempotencyKey: input.IdempotencyKey, RequestHash: requestHash,
		TargetType: input.TargetType, TargetID: input.TargetID,
		FileType: input.FileType, Title: input.Title,
	})
	if err != nil {
		return domain.FileReference{}, err
	}
	if operation.RequestHash != requestHash {
		return domain.FileReference{}, domain.ErrConflict
	}
	if operation.Status == "completed" && operation.ReferenceID != nil {
		return service.repo.GetReference(ctx, input.WorkspaceID, input.UserID, *operation.ReferenceID)
	}
	if !created {
		now := service.now().UTC()
		switch operation.Status {
		case "pending":
			if now.Before(operation.UpdatedAt.Add(createOperationLease)) {
				return domain.FileReference{}, domain.ErrOperationInProgress
			}
		case "failed":
			// Failed operations are immediately retryable. Recovery still checks
			// Google's appProperties before attempting another create.
		default:
			return domain.FileReference{}, domain.ErrConflict
		}
		operation, created, err = service.repo.ClaimOperation(
			ctx, operation.ID, operation.UpdatedAt, now.Add(-createOperationLease),
		)
		if err != nil {
			return domain.FileReference{}, err
		}
		if !created {
			return domain.FileReference{}, domain.ErrOperationInProgress
		}
	}
	connection, token, err := service.connectionToken(ctx, input.WorkspaceID, input.UserID)
	if err != nil {
		service.failCreateOperation(ctx, operation.ID, "connection_unavailable")
		return domain.FileReference{}, err
	}
	providerFile, err := service.recoverOrCreateFile(ctx, token.AccessToken, operation, input)
	if err != nil {
		service.failCreateOperation(ctx, operation.ID, "provider_create_failed")
		return domain.FileReference{}, service.handleProviderAuthorizationError(ctx, connection, err)
	}
	reference, err := service.repo.AttachFile(ctx, input.WorkspaceID, input.UserID, connection.ID, input.TargetType, input.TargetID, providerFile)
	if err != nil {
		service.failCreateOperation(ctx, operation.ID, "reference_attach_failed")
		return domain.FileReference{}, err
	}
	if err := service.repo.CompleteOperation(ctx, operation.ID, providerFile.ID, reference.ID); err != nil && !errors.Is(err, domain.ErrConflict) {
		return domain.FileReference{}, err
	}
	return reference, nil
}

func (service *Service) failCreateOperation(ctx context.Context, operationID uuid.UUID, errorCode string) {
	if err := service.repo.FailOperation(ctx, operationID, errorCode); err != nil &&
		!errors.Is(err, domain.ErrConflict) && service.log != nil {
		service.log.Error(ctx, "failed recording Google Drive create operation failure", "error", err, "operation_id", operationID)
	}
}

func (service *Service) recoverOrCreateFile(ctx context.Context, accessToken string, operation domain.CreateOperation, input CreateFileInput) (domain.ProviderFile, error) {
	if existing, err := service.client.FindCreatedFile(ctx, accessToken, operation.ID.String()); err != nil {
		return domain.ProviderFile{}, err
	} else if existing != nil {
		if err := validateCreatedFile(*existing, input.FileType); err != nil {
			return domain.ProviderFile{}, err
		}
		return *existing, nil
	}
	providerFile, err := service.client.CreateFile(ctx, accessToken, input.FileType, input.Title, operation.ID.String())
	if err != nil {
		return domain.ProviderFile{}, err
	}
	if err := validateCreatedFile(providerFile, input.FileType); err != nil {
		return domain.ProviderFile{}, err
	}
	sourceURL := workspaceurl.Build(service.config.WebsiteURL, input.WorkspaceSlug, targetRoute(input.TargetType, input.TargetID)...)
	if err := service.client.PopulateFile(ctx, accessToken, providerFile, sourceURL); err != nil && service.log != nil {
		service.log.Warn(ctx, "created Google file but could not populate backlink", "error", err, "operation_id", operation.ID)
	}
	return providerFile, nil
}

func validateTarget(workspaceID, userID uuid.UUID, targetType domain.TargetType, targetID uuid.UUID) error {
	if workspaceID == uuid.Nil || userID == uuid.Nil || targetID == uuid.Nil || !targetType.Valid() {
		return domain.ErrInvalidInput
	}
	return nil
}

func createRequestHash(input CreateFileInput) string {
	payload := strings.Join([]string{string(input.TargetType), input.TargetID.String(), string(input.FileType), input.Title}, "\x00")
	digest := sha256.Sum256([]byte(payload))
	return hex.EncodeToString(digest[:])
}

func revocationToken(token domain.OAuthToken) string {
	if refreshToken := strings.TrimSpace(token.RefreshToken); refreshToken != "" {
		return refreshToken
	}
	return token.AccessToken
}

func targetRoute(targetType domain.TargetType, targetID uuid.UUID) []string {
	switch targetType {
	case domain.TargetStory:
		return []string{"story", targetID.String()}
	case domain.TargetDocument:
		return []string{"docs", targetID.String()}
	default:
		// Objective routes include their team ID, which this provider boundary
		// intentionally does not infer. Use the valid workspace root instead of
		// writing a broken deep link into the newly created Google file.
		return []string{}
	}
}

func (service *Service) logProviderError(ctx context.Context, message string, err error, workspaceID, userID uuid.UUID) {
	if service.log == nil {
		return
	}
	service.log.Error(ctx, message, "error", err, "workspace_id", workspaceID, "user_id", userID)
}

func (service *Service) handleProviderAuthorizationError(ctx context.Context, account domain.Account, err error) error {
	var apiError *APIError
	if !errors.As(err, &apiError) || apiError.StatusCode != http.StatusUnauthorized {
		return err
	}
	_ = service.repo.MarkReauthorizationRequired(ctx, account, "provider_unauthorized")
	return domain.ErrReauthorizationRequired
}

func (service *Service) validatedWorkspaceOrigin(workspaceSlug, raw string) string {
	parsed, err := url.Parse(strings.TrimSpace(raw))
	if err != nil || parsed.User != nil || (parsed.Scheme != "http" && parsed.Scheme != "https") || parsed.Hostname() == "" {
		return ""
	}
	expected, err := url.Parse(workspaceurl.Build(service.config.WebsiteURL, workspaceSlug))
	if err != nil || expected.Scheme != parsed.Scheme || !strings.EqualFold(expected.Host, parsed.Host) {
		return ""
	}
	return parsed.Scheme + "://" + parsed.Host
}

func (service *Service) workspaceOrigin(workspaceSlug string) string {
	parsed, err := url.Parse(workspaceurl.Build(service.config.WebsiteURL, workspaceSlug))
	if err != nil || parsed.User != nil || (parsed.Scheme != "http" && parsed.Scheme != "https") || parsed.Hostname() == "" {
		return ""
	}
	return parsed.Scheme + "://" + parsed.Host
}

func appendQuery(rawURL, key, value string) string {
	parsed, err := url.Parse(rawURL)
	if err != nil || parsed.Host == "" {
		return "/"
	}
	query := parsed.Query()
	query.Set(key, value)
	parsed.RawQuery = query.Encode()
	return parsed.String()
}

func clearBytes(value []byte) {
	for index := range value {
		value[index] = 0
	}
}

func randomURLToken(size int) (string, error) {
	value := make([]byte, size)
	if _, err := rand.Read(value); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(value), nil
}

func marshalToken(token domain.OAuthToken) ([]byte, error) {
	return json.Marshal(token)
}
