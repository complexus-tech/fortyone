package googledrive

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/complexus-tech/projects-api/internal/modules/googledrive/domain"
	"github.com/complexus-tech/projects-api/internal/platform/credentialvault"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestUnconfiguredIntegrationFailsClosed(t *testing.T) {
	t.Parallel()

	service := New(nil, nil, Config{})
	require.False(t, service.configured())
	_, err := service.CreateConnectSession(t.Context(), uuid.New(), uuid.New(), "acme", "")
	require.ErrorContains(t, err, "not configured")
	_, err = service.PickerSession(t.Context(), uuid.New(), uuid.New(), "acme", "https://acme.fortyone.app")
	require.ErrorContains(t, err, "not configured")
	_, err = service.AttachFiles(t.Context(), AttachInput{})
	require.ErrorIs(t, err, domain.ErrNotConfigured)
	_, err = service.RefreshFile(t.Context(), uuid.New(), uuid.New(), uuid.New())
	require.ErrorIs(t, err, domain.ErrNotConfigured)
	_, err = service.ReadContent(t.Context(), uuid.New(), uuid.New(), uuid.New())
	require.ErrorIs(t, err, domain.ErrNotConfigured)
	_, err = service.ReadPreview(t.Context(), uuid.New(), uuid.New(), uuid.New())
	require.ErrorIs(t, err, domain.ErrNotConfigured)
	_, err = service.CreateFile(t.Context(), CreateFileInput{})
	require.ErrorIs(t, err, domain.ErrNotConfigured)
	_, err = service.ImportFile(t.Context(), ImportInput{})
	require.ErrorIs(t, err, domain.ErrNotConfigured)
}

func TestDisconnectUsesUserProviderGateWithoutCallingGoogle(t *testing.T) {
	t.Parallel()

	events := make([]string, 0, 3)
	repo := &oauthRepositoryStub{events: &events}
	client := &providerClientStub{}
	service := &Service{repo: repo, client: client}

	err := service.Disconnect(t.Context(), uuid.New(), uuid.New())

	require.NoError(t, err)
	require.Equal(t, []string{"user_lock_begin", "disconnect", "user_lock_end"}, events)
	require.Equal(t, 1, repo.disconnectCalls)
	require.Zero(t, client.revokeCalls)
}

func TestPickerOriginMustMatchWorkspaceOrigin(t *testing.T) {
	t.Parallel()

	service := &Service{config: Config{WebsiteURL: "https://fortyone.app"}}
	require.Equal(t, "https://acme.fortyone.app", service.validatedWorkspaceOrigin("acme", "https://acme.fortyone.app"))
	require.Empty(t, service.validatedWorkspaceOrigin("acme", "https://attacker.example"))
	require.Empty(t, service.validatedWorkspaceOrigin("acme", "javascript:alert(1)"))
	require.Equal(t, "https://acme.fortyone.app", service.workspaceOrigin("acme"))
}

func TestTargetRouteUsesValidFortyOnePaths(t *testing.T) {
	t.Parallel()

	targetID := uuid.New()
	require.Equal(t, []string{"story", targetID.String()}, targetRoute(domain.TargetStory, targetID))
	require.Equal(t, []string{"docs", targetID.String()}, targetRoute(domain.TargetDocument, targetID))
	require.Empty(t, targetRoute(domain.TargetObjective, targetID))
	require.Empty(t, targetRoute(domain.TargetComment, targetID))
}

func TestFileReferenceJSONKeepsProviderCredentialsInternal(t *testing.T) {
	t.Parallel()

	referenceID := uuid.New()
	data, err := json.Marshal(domain.FileReference{
		ID: referenceID, FileID: "provider-file-secret", ResourceKey: pointer("resource-key-secret"),
		Name: "Launch brief", MimeType: googleDocumentMimeType,
	})
	require.NoError(t, err)
	require.Contains(t, string(data), referenceID.String())
	require.NotContains(t, string(data), "provider-file-secret")
	require.NotContains(t, string(data), "resource-key-secret")
}

func TestCreateRecoveryLooksUpProviderBeforeCreating(t *testing.T) {
	t.Parallel()

	existing := domain.ProviderFile{ID: "provider-file", Name: "Recovered", MimeType: googleDocumentMimeType, WebViewLink: "https://docs.google.com/document/d/provider-file/edit"}
	client := &providerClientStub{findResult: &existing}
	service := &Service{client: client}
	now := time.Now().UTC()
	operation := domain.CreateOperation{ID: uuid.New(), CreatedAt: now, UpdatedAt: now}

	result, err := service.recoverOrCreateFile(t.Context(), "access-token", operation, CreateFileInput{
		FileType: domain.FileTypeDocument, Title: "Plan",
	})
	require.NoError(t, err)
	require.Equal(t, existing, result)
	require.Equal(t, 1, client.findCalls)
	require.Zero(t, client.createCalls)
}

func TestReadFileIsBoundedAndReportsTruncation(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		require.Equal(t, "Bearer access-token", request.Header.Get("Authorization"))
		_, _ = writer.Write([]byte("abcdef"))
	}))
	t.Cleanup(server.Close)
	client := newGoogleClient(server.Client(), Config{}, nil)
	client.driveURL = server.URL

	content, err := client.ReadFile(t.Context(), "access-token", domain.ProviderFile{
		ID: "provider-file", MimeType: "text/plain",
	}, 4)
	require.NoError(t, err)
	require.Equal(t, "abcd", content.Text)
	require.True(t, content.Truncated)
	require.Equal(t, 4, content.BytesRead)
}

func TestReadFilePreservesValidUTF8AtByteLimit(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
		_, _ = writer.Write([]byte("abc界z"))
	}))
	t.Cleanup(server.Close)
	client := newGoogleClient(server.Client(), Config{}, nil)
	client.driveURL = server.URL

	content, err := client.ReadFile(t.Context(), "access-token", domain.ProviderFile{
		ID: "provider-file", MimeType: "text/plain",
	}, 5)
	require.NoError(t, err)
	require.Equal(t, "abc", content.Text)
	require.True(t, content.Truncated)
	require.Equal(t, 3, content.BytesRead)
}

func TestReadPreviewUsesTheActorGrantAndRevalidatesBeforeProxying(t *testing.T) {
	t.Parallel()

	workspaceID := uuid.New()
	userID := uuid.New()
	accountID := uuid.New()
	referenceID := uuid.New()
	revalidatedGeneration := uuid.New()
	account := domain.Account{
		ID: accountID, UserID: userID, GoogleSubject: "google-subject",
		CredentialVersion:      int16(credentialvault.CurrentVersion),
		InstallationGeneration: uuid.New(), ExpiresAt: time.Now().Add(time.Hour),
	}
	repo := &referenceRepositoryStub{revalidatedGrantGeneration: revalidatedGeneration}
	client := &providerClientStub{
		getFileResult: domain.ProviderFile{
			ID: "provider-file", Name: "Launch plan", MimeType: googleDocumentMimeType,
			WebViewLink:   "https://docs.google.com/document/d/provider-file/edit",
			ThumbnailLink: "https://lh3.googleusercontent.com/thumbnail",
		},
		thumbnailResult: Preview{Bytes: []byte("preview"), ContentType: "image/png"},
	}
	service := New(nil, repo, configuredOAuthTestConfig(t))
	service.client = client
	payload, err := service.sealToken(account, domain.OAuthToken{
		AccessToken: "access-token", RefreshToken: "refresh-token", Expiry: account.ExpiresAt,
	})
	require.NoError(t, err)
	account.CredentialPayload = payload
	repo.reference = domain.FileReference{
		ID: referenceID, FileID: "provider-file", Name: "Launch plan",
		MimeType: googleDocumentMimeType, TargetType: domain.TargetStory,
		TargetID: uuid.New(), Account: &account, GrantGeneration: pointerUUID(uuid.New()),
	}

	preview, err := service.ReadPreview(t.Context(), workspaceID, userID, referenceID)

	require.NoError(t, err)
	require.Equal(t, []byte("preview"), preview.Bytes)
	require.Equal(t, "image/png", preview.ContentType)
	require.Equal(t, workspaceID, repo.getReferenceWorkspaceID)
	require.Equal(t, userID, repo.getReferenceUserID)
	require.Equal(t, referenceID, repo.getReferenceID)
	require.Equal(t, 1, repo.revalidateCalls)
	require.Equal(t, workspaceID, repo.revalidatedWorkspaceID)
	require.Equal(t, userID, repo.revalidatedUserID)
	require.Equal(t, accountID, repo.revalidatedAccountID)
	require.Equal(t, 1, client.getFileCalls)
	require.Equal(t, 1, client.thumbnailCalls)
	require.Equal(t, "access-token", client.thumbnailToken)
	require.Equal(t, "https://lh3.googleusercontent.com/thumbnail", client.thumbnailURL)
	require.Equal(t, maxPreviewBytes, client.thumbnailLimit)
}

func TestReadPreviewRejectsAReferenceGrantOwnedByAnotherUser(t *testing.T) {
	t.Parallel()

	workspaceID := uuid.New()
	userID := uuid.New()
	referenceID := uuid.New()
	repo := &referenceRepositoryStub{
		reference: domain.FileReference{
			ID: referenceID, FileID: "provider-file",
			Account: &domain.Account{ID: uuid.New(), UserID: uuid.New()},
		},
	}
	client := &providerClientStub{}
	service := New(nil, repo, configuredOAuthTestConfig(t))
	service.client = client

	_, err := service.ReadPreview(t.Context(), workspaceID, userID, referenceID)

	require.ErrorIs(t, err, domain.ErrForbidden)
	require.Equal(t, userID, repo.getReferenceUserID)
	require.Zero(t, client.getFileCalls)
	require.Zero(t, client.thumbnailCalls)
	require.Zero(t, repo.revalidateCalls)
}

func TestPreviewAccessLossInvalidatesTheRevalidatedGrantGeneration(t *testing.T) {
	t.Parallel()

	workspaceID := uuid.New()
	userID := uuid.New()
	accountID := uuid.New()
	referenceID := uuid.New()
	staleGeneration := uuid.New()
	revalidatedGeneration := uuid.New()
	account := domain.Account{
		ID: accountID, UserID: userID, GoogleSubject: "google-subject",
		CredentialVersion:      int16(credentialvault.CurrentVersion),
		InstallationGeneration: uuid.New(), ExpiresAt: time.Now().Add(time.Hour),
	}
	repo := &referenceRepositoryStub{revalidatedGrantGeneration: revalidatedGeneration}
	providerErr := &APIError{
		StatusCode: http.StatusForbidden,
		Reasons:    []string{"insufficientFilePermissions"},
	}
	client := &providerClientStub{
		getFileResult: domain.ProviderFile{
			ID: "provider-file", Name: "Launch plan", MimeType: googleDocumentMimeType,
			ThumbnailLink: "https://lh3.googleusercontent.com/thumbnail",
		},
		thumbnailErr: providerErr,
	}
	service := New(nil, repo, configuredOAuthTestConfig(t))
	service.client = client
	payload, err := service.sealToken(account, domain.OAuthToken{
		AccessToken: "access-token", RefreshToken: "refresh-token", Expiry: account.ExpiresAt,
	})
	require.NoError(t, err)
	account.CredentialPayload = payload
	repo.reference = domain.FileReference{
		ID: referenceID, FileID: "provider-file", Name: "Launch plan",
		MimeType: googleDocumentMimeType, TargetType: domain.TargetStory,
		TargetID: uuid.New(), Account: &account, GrantGeneration: &staleGeneration,
	}

	_, previewErr := service.ReadPreview(t.Context(), workspaceID, userID, referenceID)

	require.Same(t, providerErr, previewErr)
	require.Equal(t, 1, repo.revalidateCalls)
	require.Equal(t, 1, client.thumbnailCalls)
	require.Equal(t, 1, repo.deleteGrantCalls)
	require.Equal(t, revalidatedGeneration, repo.deletedGrantGeneration)
	require.NotEqual(t, staleGeneration, repo.deletedGrantGeneration)
}

func TestParseAPIErrorPreservesDriveReasonsAndClassifiesRateLimits(t *testing.T) {
	t.Parallel()

	recorder := httptest.NewRecorder()
	recorder.Code = http.StatusForbidden
	_, _ = recorder.Body.WriteString(`{"error":{"code":403,"message":"quota","status":"PERMISSION_DENIED","errors":[{"reason":"userRateLimitExceeded"}]}}`)
	response := recorder.Result()
	t.Cleanup(func() { _ = response.Body.Close() })

	var apiError *APIError
	require.ErrorAs(t, parseAPIError(response), &apiError)
	require.Equal(t, http.StatusForbidden, apiError.StatusCode)
	require.Equal(t, "permission_denied", apiError.Code)
	require.Equal(t, []string{"userRateLimitExceeded"}, apiError.Reasons)
	require.True(t, apiError.IsRateLimited())
	require.False(t, apiError.isFilePermissionLoss())
}

func TestNormalizedImportTitleIsBoundedAndHasFallback(t *testing.T) {
	t.Parallel()

	require.Equal(t, "Imported Google Doc", normalizedImportTitle("  "))
	title := strings.Repeat("界", maxTitleRunes+10)
	require.Len(t, []rune(normalizedImportTitle(title)), maxTitleRunes)
}

func TestRevocationTokenPrefersRefreshToken(t *testing.T) {
	t.Parallel()

	require.Equal(t, "refresh", revocationToken(domain.OAuthToken{AccessToken: "access", RefreshToken: " refresh "}))
	require.Equal(t, "access", revocationToken(domain.OAuthToken{AccessToken: "access"}))
}

func TestNormalizeResourceKeyRejectsHeaderDelimitersAndControls(t *testing.T) {
	t.Parallel()

	valid := " resource-key_123 "
	normalized, err := normalizeResourceKey(&valid)
	require.NoError(t, err)
	require.Equal(t, "resource-key_123", *normalized)
	for _, invalid := range []string{"key,other", "file/key", "key\r\nheader", strings.Repeat("a", maxResourceKeyBytes+1)} {
		_, err := normalizeResourceKey(&invalid)
		require.ErrorIs(t, err, domain.ErrInvalidInput)
	}
}

func TestReferenceAccessLossInvalidatesOnlyTheActorGrant(t *testing.T) {
	t.Parallel()

	workspaceID := uuid.New()
	userID := uuid.New()
	accountID := uuid.New()
	referenceID := uuid.New()
	grantGeneration := uuid.New()
	repo := &referenceRepositoryStub{}
	service := &Service{repo: repo}
	providerErr := &APIError{StatusCode: http.StatusForbidden, Reasons: []string{"insufficientFilePermissions"}}

	err := service.handleReferenceProviderError(
		t.Context(),
		workspaceID,
		userID,
		domain.FileReference{ID: referenceID, GrantGeneration: &grantGeneration},
		domain.Account{ID: accountID},
		providerErr,
	)

	require.Same(t, providerErr, err)
	require.Equal(t, 1, repo.deleteGrantCalls)
	require.Equal(t, workspaceID, repo.deletedGrantWorkspaceID)
	require.Equal(t, userID, repo.deletedGrantUserID)
	require.Equal(t, accountID, repo.deletedGrantAccountID)
	require.Equal(t, referenceID, repo.deletedGrantReferenceID)
	require.Equal(t, grantGeneration, repo.deletedGrantGeneration)
}

func TestExportAccessLossInvalidatesTheRevalidatedGrantGeneration(t *testing.T) {
	t.Parallel()

	for _, test := range []struct {
		name   string
		export func(context.Context, *Service, uuid.UUID, uuid.UUID, uuid.UUID) error
	}{
		{
			name: "read content",
			export: func(ctx context.Context, service *Service, workspaceID, userID, referenceID uuid.UUID) error {
				_, err := service.ReadContent(ctx, workspaceID, userID, referenceID)
				return err
			},
		},
		{
			name: "import file",
			export: func(ctx context.Context, service *Service, workspaceID, userID, referenceID uuid.UUID) error {
				_, err := service.ImportFile(ctx, ImportInput{
					WorkspaceID: workspaceID, UserID: userID,
					ReferenceID: referenceID, Visibility: "private",
					IdempotencyKey: "import-operation-1",
				})
				return err
			},
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			workspaceID := uuid.New()
			userID := uuid.New()
			accountID := uuid.New()
			referenceID := uuid.New()
			staleGeneration := uuid.New()
			revalidatedGeneration := uuid.New()
			account := domain.Account{
				ID: accountID, UserID: userID, GoogleSubject: "google-subject",
				CredentialVersion:      int16(credentialvault.CurrentVersion),
				InstallationGeneration: uuid.New(), ExpiresAt: time.Now().Add(time.Hour),
			}
			reference := domain.FileReference{
				ID: referenceID, FileID: "provider-file", Name: "Launch plan",
				MimeType: googleDocumentMimeType, TargetType: domain.TargetStory,
				TargetID: uuid.New(), Account: &account, GrantGeneration: &staleGeneration,
			}
			repo := &referenceRepositoryStub{
				reference: reference, targetMutable: true,
				revalidatedGrantGeneration: revalidatedGeneration,
			}
			providerErr := &APIError{
				StatusCode: http.StatusForbidden,
				Reasons:    []string{"insufficientFilePermissions"},
			}
			client := &providerClientStub{
				getFileResult: domain.ProviderFile{
					ID: "provider-file", Name: "Launch plan", MimeType: googleDocumentMimeType,
					WebViewLink: "https://docs.google.com/document/d/provider-file/edit",
				},
				readFileErr: providerErr,
			}
			vault, err := credentialvault.NewFromEncodedKeyring(
				credentialvault.DevelopmentKeyID,
				credentialvault.DevelopmentKeyVersion,
				credentialvault.DevelopmentEncodedKeys,
			)
			require.NoError(t, err)
			service := New(nil, repo, Config{
				ClientID: "client", ClientSecret: "secret", RedirectURL: "https://example.com/callback",
				PickerAPIKey: "picker", AppID: "123", Credentials: vault,
			})
			service.client = client
			account.CredentialPayload, err = service.sealToken(account, domain.OAuthToken{
				AccessToken: "access-token", RefreshToken: "refresh-token", Expiry: account.ExpiresAt,
			})
			require.NoError(t, err)

			exportErr := test.export(t.Context(), service, workspaceID, userID, referenceID)

			require.Same(t, providerErr, exportErr)
			require.Equal(t, 1, repo.revalidateCalls)
			require.Equal(t, 1, repo.deleteGrantCalls)
			require.Equal(t, revalidatedGeneration, repo.deletedGrantGeneration)
			require.NotEqual(t, staleGeneration, repo.deletedGrantGeneration)
			require.Equal(t, 1, client.readFileCalls)
		})
	}
}

func TestTransientForbiddenPreservesTheActorGrant(t *testing.T) {
	t.Parallel()

	repo := &referenceRepositoryStub{}
	service := &Service{repo: repo}
	providerErr := &APIError{StatusCode: http.StatusForbidden, Reasons: []string{"userRateLimitExceeded"}}

	err := service.handleReferenceProviderError(
		t.Context(), uuid.New(), uuid.New(),
		domain.FileReference{ID: uuid.New(), GrantGeneration: pointerUUID(uuid.New())},
		domain.Account{ID: uuid.New()}, providerErr,
	)

	require.Same(t, providerErr, err)
	require.Zero(t, repo.deleteGrantCalls)
}

func TestConfirmedUnavailableFilePersistsItsState(t *testing.T) {
	t.Parallel()

	workspaceID := uuid.New()
	userID := uuid.New()
	referenceID := uuid.New()
	repo := &referenceRepositoryStub{}
	service := &Service{repo: repo}

	err := service.markReferenceUnavailable(t.Context(), workspaceID, userID, referenceID)

	require.ErrorIs(t, err, domain.ErrNotFound)
	require.Equal(t, 1, repo.markUnavailableCalls)
	require.Equal(t, workspaceID, repo.unavailableWorkspaceID)
	require.Equal(t, referenceID, repo.unavailableReferenceID)
}
