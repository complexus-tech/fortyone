package googledrivehttp

import (
	"context"
	"errors"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/complexus-tech/projects-api/internal/modules/googledrive/domain"
	googledrive "github.com/complexus-tech/projects-api/internal/modules/googledrive/service"
	mid "github.com/complexus-tech/projects-api/internal/platform/http/middleware"
	"github.com/complexus-tech/projects-api/pkg/web"
	"github.com/google/uuid"
)

const googleDriveJSONBodyLimit int64 = 32 << 10

type Handlers struct{ service *googledrive.Service }

func New(service *googledrive.Service) *Handlers { return &Handlers{service: service} }

type connectInput struct {
	ReturnURL string `json:"returnUrl"`
}

type selectedFileInput struct {
	ID          string  `json:"id" validate:"required"`
	Name        *string `json:"name"`
	MimeType    *string `json:"mimeType"`
	ResourceKey *string `json:"resourceKey"`
}

type targetInput struct {
	TargetType string    `json:"targetType" validate:"required"`
	TargetID   uuid.UUID `json:"targetId" validate:"required"`
}

type attachInput struct {
	TargetType string              `json:"targetType" validate:"required"`
	TargetID   uuid.UUID           `json:"targetId" validate:"required"`
	Files      []selectedFileInput `json:"files" validate:"required,min=1,max=20,dive"`
}

type createFileInput struct {
	TargetType string    `json:"targetType" validate:"required"`
	TargetID   uuid.UUID `json:"targetId" validate:"required"`
	FileType   string    `json:"fileType" validate:"required,oneof=document spreadsheet"`
	Title      string    `json:"title" validate:"required,max=1024"`
}

type importFileInput struct {
	Visibility string `json:"visibility" validate:"required,oneof=workspace private"`
}

func (handlers *Handlers) GetIntegration(ctx context.Context, writer http.ResponseWriter, _ *http.Request) error {
	writer.Header().Set("Cache-Control", "private, no-store")
	workspace, userID, err := requestContext(ctx)
	if err != nil {
		return web.RespondError(ctx, writer, err, http.StatusUnauthorized)
	}
	result, err := handlers.service.GetIntegration(ctx, workspace.ID, userID)
	if err != nil {
		return web.RespondError(ctx, writer, err, googleDriveHTTPStatus(err))
	}
	return web.Respond(ctx, writer, result, http.StatusOK)
}

func (handlers *Handlers) CreateConnectSession(ctx context.Context, writer http.ResponseWriter, request *http.Request) error {
	workspace, userID, err := requestContext(ctx)
	if err != nil {
		return web.RespondError(ctx, writer, err, http.StatusUnauthorized)
	}
	var input connectInput
	if err := web.DecodeWithLimit(request, &input, googleDriveJSONBodyLimit); err != nil {
		return web.RespondError(ctx, writer, err, http.StatusBadRequest)
	}
	authorizationURL, err := handlers.service.CreateConnectSession(ctx, workspace.ID, userID, workspace.Slug, input.ReturnURL)
	if err != nil {
		return web.RespondError(ctx, writer, err, googleDriveHTTPStatus(err))
	}
	writer.Header().Set("Cache-Control", "private, no-store")
	return web.Respond(ctx, writer, map[string]string{"authorizationUrl": authorizationURL}, http.StatusOK)
}

func (handlers *Handlers) CompleteOAuth(ctx context.Context, writer http.ResponseWriter, request *http.Request) error {
	writer.Header().Set("Cache-Control", "private, no-store")
	userID, err := mid.GetUserID(ctx)
	if err != nil {
		return web.RespondError(ctx, writer, err, http.StatusUnauthorized)
	}
	callback, err := web.ParseOAuthCallbackQuery(request.URL.Query())
	if err != nil {
		return web.RespondError(ctx, writer, err, http.StatusBadRequest)
	}
	redirectURL, err := handlers.service.CompleteOAuth(ctx, userID, callback.Code, callback.State, callback.ProviderError)
	if err != nil {
		if redirectURL != "" {
			// #nosec G710 -- redirectURL is reconstructed from a consumed server-side
			// OAuth state and constrained to the configured workspace origin.
			http.Redirect(writer, request, redirectURL, http.StatusTemporaryRedirect)
			return nil
		}
		return web.RespondError(ctx, writer, err, googleDriveHTTPStatus(err))
	}
	// #nosec G710 -- redirectURL is reconstructed from a consumed server-side
	// OAuth state and constrained to the configured workspace origin.
	http.Redirect(writer, request, redirectURL, http.StatusTemporaryRedirect)
	return nil
}

func (handlers *Handlers) Disconnect(ctx context.Context, writer http.ResponseWriter, _ *http.Request) error {
	workspace, userID, err := requestContext(ctx)
	if err != nil {
		return web.RespondError(ctx, writer, err, http.StatusUnauthorized)
	}
	if err := handlers.service.Disconnect(ctx, workspace.ID, userID); err != nil {
		return web.RespondError(ctx, writer, err, googleDriveHTTPStatus(err))
	}
	return web.Respond(ctx, writer, nil, http.StatusNoContent)
}

func (handlers *Handlers) CreatePickerSession(ctx context.Context, writer http.ResponseWriter, request *http.Request) error {
	workspace, userID, err := requestContext(ctx)
	if err != nil {
		return web.RespondError(ctx, writer, err, http.StatusUnauthorized)
	}
	result, err := handlers.service.PickerSession(ctx, workspace.ID, userID, workspace.Slug, request.Header.Get("Origin"))
	if err != nil {
		return web.RespondError(ctx, writer, err, googleDriveHTTPStatus(err))
	}
	writer.Header().Set("Cache-Control", "private, no-store")
	return web.Respond(ctx, writer, result, http.StatusOK)
}

func (handlers *Handlers) ListFiles(ctx context.Context, writer http.ResponseWriter, request *http.Request) error {
	writer.Header().Set("Cache-Control", "private, no-store")
	workspace, userID, err := requestContext(ctx)
	if err != nil {
		return web.RespondError(ctx, writer, err, http.StatusUnauthorized)
	}
	target, err := targetFromQuery(request)
	if err != nil {
		return web.RespondError(ctx, writer, err, http.StatusBadRequest)
	}
	result, err := handlers.service.ListFiles(ctx, workspace.ID, userID, domain.TargetType(target.TargetType), target.TargetID)
	if err != nil {
		return web.RespondError(ctx, writer, err, googleDriveHTTPStatus(err))
	}
	return web.Respond(ctx, writer, result, http.StatusOK)
}

func (handlers *Handlers) AttachFiles(ctx context.Context, writer http.ResponseWriter, request *http.Request) error {
	workspace, userID, err := requestContext(ctx)
	if err != nil {
		return web.RespondError(ctx, writer, err, http.StatusUnauthorized)
	}
	var input attachInput
	if err := web.DecodeWithLimit(request, &input, googleDriveJSONBodyLimit); err != nil {
		return web.RespondError(ctx, writer, err, http.StatusBadRequest)
	}
	selected := make([]googledrive.SelectedFile, 0, len(input.Files))
	for _, file := range input.Files {
		selected = append(selected, googledrive.SelectedFile{ID: file.ID, Name: file.Name, MimeType: file.MimeType, ResourceKey: file.ResourceKey})
	}
	result, err := handlers.service.AttachFiles(ctx, googledrive.AttachInput{
		WorkspaceID: workspace.ID, UserID: userID,
		TargetType: domain.TargetType(input.TargetType), TargetID: input.TargetID, Files: selected,
	})
	if err != nil {
		return web.RespondError(ctx, writer, err, googleDriveHTTPStatus(err))
	}
	return web.Respond(ctx, writer, result, http.StatusCreated)
}

func (handlers *Handlers) CreateFile(ctx context.Context, writer http.ResponseWriter, request *http.Request) error {
	workspace, userID, err := requestContext(ctx)
	if err != nil {
		return web.RespondError(ctx, writer, err, http.StatusUnauthorized)
	}
	var input createFileInput
	if err := web.DecodeWithLimit(request, &input, googleDriveJSONBodyLimit); err != nil {
		return web.RespondError(ctx, writer, err, http.StatusBadRequest)
	}
	idempotencyKey := strings.TrimSpace(request.Header.Get("Idempotency-Key"))
	if idempotencyKey == "" {
		return web.RespondError(ctx, writer, domain.ErrInvalidInput, http.StatusBadRequest)
	}
	result, err := handlers.service.CreateFile(ctx, googledrive.CreateFileInput{
		WorkspaceID: workspace.ID, UserID: userID, WorkspaceSlug: workspace.Slug,
		TargetType: domain.TargetType(input.TargetType), TargetID: input.TargetID,
		FileType: domain.FileType(input.FileType), Title: input.Title,
		IdempotencyKey: idempotencyKey,
	})
	if err != nil {
		return web.RespondError(ctx, writer, err, googleDriveHTTPStatus(err))
	}
	return web.Respond(ctx, writer, result, http.StatusCreated)
}

func (handlers *Handlers) ReadContent(ctx context.Context, writer http.ResponseWriter, request *http.Request) error {
	writer.Header().Set("Cache-Control", "private, no-store")
	workspace, userID, referenceID, err := referenceContext(ctx, request)
	if err != nil {
		return web.RespondError(ctx, writer, err, http.StatusBadRequest)
	}
	result, err := handlers.service.ReadContent(ctx, workspace.ID, userID, referenceID)
	if err != nil {
		return web.RespondError(ctx, writer, err, googleDriveHTTPStatus(err))
	}
	return web.Respond(ctx, writer, result, http.StatusOK)
}

func (handlers *Handlers) ReadPreview(ctx context.Context, writer http.ResponseWriter, request *http.Request) error {
	writer.Header().Set("Cache-Control", "private, no-store")
	workspace, userID, referenceID, err := referenceContext(ctx, request)
	if err != nil {
		return web.RespondError(ctx, writer, err, http.StatusBadRequest)
	}
	result, err := handlers.service.ReadPreview(ctx, workspace.ID, userID, referenceID)
	if err != nil {
		return web.RespondError(ctx, writer, err, googleDriveHTTPStatus(err))
	}
	return writePreviewResponse(writer, result)
}

func writePreviewResponse(writer http.ResponseWriter, preview googledrive.Preview) error {
	switch preview.ContentType {
	case "image/jpeg", "image/png", "image/webp":
	default:
		return errors.New("Google Drive preview has an unsupported content type")
	}
	if len(preview.Bytes) == 0 {
		return errors.New("Google Drive preview is empty")
	}
	writer.Header().Set("Cache-Control", "private, no-store")
	writer.Header().Set("Content-Type", preview.ContentType)
	writer.Header().Set("Content-Length", strconv.Itoa(len(preview.Bytes)))
	writer.Header().Set("Content-Security-Policy", "default-src 'none'; sandbox")
	writer.Header().Set("Cross-Origin-Resource-Policy", "same-site")
	writer.Header().Set("Referrer-Policy", "no-referrer")
	writer.Header().Set("X-Content-Type-Options", "nosniff")
	writer.WriteHeader(http.StatusOK)
	written, err := writer.Write(preview.Bytes)
	if err != nil {
		return err
	}
	if written != len(preview.Bytes) {
		return io.ErrShortWrite
	}
	return nil
}

func (handlers *Handlers) RefreshFile(ctx context.Context, writer http.ResponseWriter, request *http.Request) error {
	workspace, userID, referenceID, err := referenceContext(ctx, request)
	if err != nil {
		return web.RespondError(ctx, writer, err, http.StatusBadRequest)
	}
	result, err := handlers.service.RefreshFile(ctx, workspace.ID, userID, referenceID)
	if err != nil {
		return web.RespondError(ctx, writer, err, googleDriveHTTPStatus(err))
	}
	return web.Respond(ctx, writer, result, http.StatusOK)
}

func (handlers *Handlers) DeleteFile(ctx context.Context, writer http.ResponseWriter, request *http.Request) error {
	workspace, userID, referenceID, err := referenceContext(ctx, request)
	if err != nil {
		return web.RespondError(ctx, writer, err, http.StatusBadRequest)
	}
	if err := handlers.service.DeleteFile(ctx, workspace.ID, userID, referenceID); err != nil {
		return web.RespondError(ctx, writer, err, googleDriveHTTPStatus(err))
	}
	return web.Respond(ctx, writer, nil, http.StatusNoContent)
}

func (handlers *Handlers) ImportFile(ctx context.Context, writer http.ResponseWriter, request *http.Request) error {
	workspace, userID, referenceID, err := referenceContext(ctx, request)
	if err != nil {
		return web.RespondError(ctx, writer, err, http.StatusBadRequest)
	}
	var input importFileInput
	if err := web.DecodeWithLimit(request, &input, googleDriveJSONBodyLimit); err != nil {
		return web.RespondError(ctx, writer, err, http.StatusBadRequest)
	}
	idempotencyKey := strings.TrimSpace(request.Header.Get("Idempotency-Key"))
	if idempotencyKey == "" {
		return web.RespondError(ctx, writer, domain.ErrInvalidInput, http.StatusBadRequest)
	}
	result, err := handlers.service.ImportFile(ctx, googledrive.ImportInput{
		WorkspaceID: workspace.ID, UserID: userID, ReferenceID: referenceID,
		Visibility: input.Visibility, IdempotencyKey: idempotencyKey,
	})
	if err != nil {
		return web.RespondError(ctx, writer, err, googleDriveHTTPStatus(err))
	}
	return web.Respond(ctx, writer, result, http.StatusCreated)
}

func requestContext(ctx context.Context) (mid.WorkspaceInfo, uuid.UUID, error) {
	workspace, err := mid.GetWorkspace(ctx)
	if err != nil {
		return mid.WorkspaceInfo{}, uuid.Nil, err
	}
	userID, err := mid.GetUserID(ctx)
	return workspace, userID, err
}

func referenceContext(ctx context.Context, request *http.Request) (mid.WorkspaceInfo, uuid.UUID, uuid.UUID, error) {
	workspace, userID, err := requestContext(ctx)
	if err != nil {
		return mid.WorkspaceInfo{}, uuid.Nil, uuid.Nil, err
	}
	referenceID, err := uuid.Parse(web.Params(request, "referenceId"))
	return workspace, userID, referenceID, err
}

func targetFromQuery(request *http.Request) (targetInput, error) {
	targetType, _, err := web.OptionalQueryParameter(request.URL.Query(), "targetType", 32)
	if err != nil {
		return targetInput{}, err
	}
	rawTargetID, _, err := web.OptionalOpaqueQueryParameter(request.URL.Query(), "targetId", 64)
	if err != nil {
		return targetInput{}, err
	}
	targetID, err := uuid.Parse(rawTargetID)
	if err != nil {
		return targetInput{}, err
	}
	return targetInput{TargetType: targetType, TargetID: targetID}, nil
}

func googleDriveHTTPStatus(err error) int {
	var apiError *googledrive.APIError
	switch {
	case errors.Is(err, domain.ErrInvalidInput):
		return http.StatusBadRequest
	case errors.Is(err, domain.ErrForbidden):
		return http.StatusForbidden
	case errors.Is(err, domain.ErrNotFound):
		return http.StatusNotFound
	case errors.Is(err, domain.ErrConflict), errors.Is(err, domain.ErrAccountOwned), errors.Is(err, domain.ErrOperationInProgress),
		errors.Is(err, domain.ErrNotConnected), errors.Is(err, domain.ErrReauthorizationRequired):
		return http.StatusConflict
	case errors.Is(err, domain.ErrNotConfigured):
		return http.StatusServiceUnavailable
	case errors.Is(err, domain.ErrContentTooLarge):
		return http.StatusRequestEntityTooLarge
	case errors.As(err, &apiError) && apiError.IsRateLimited():
		return http.StatusTooManyRequests
	case errors.As(err, &apiError) && apiError.StatusCode == http.StatusNotFound:
		return http.StatusNotFound
	case errors.As(err, &apiError) && (apiError.StatusCode == http.StatusUnauthorized || apiError.StatusCode == http.StatusForbidden):
		return http.StatusForbidden
	case errors.As(err, &apiError) && apiError.StatusCode >= http.StatusInternalServerError:
		return http.StatusBadGateway
	case errors.As(err, &apiError):
		return http.StatusBadRequest
	default:
		return http.StatusInternalServerError
	}
}
