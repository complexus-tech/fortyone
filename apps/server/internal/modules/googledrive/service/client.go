package googledrive

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/complexus-tech/projects-api/internal/modules/googledrive/domain"
)

const (
	googleAuthorizationEndpoint = "https://accounts.google.com/o/oauth2/v2/auth"
	googleTokenEndpoint         = "https://oauth2.googleapis.com/token"
	googleRevokeEndpoint        = "https://oauth2.googleapis.com/revoke"
	googleUserInfoEndpoint      = "https://openidconnect.googleapis.com/v1/userinfo"
	googleDriveAPIBase          = "https://www.googleapis.com/drive/v3"
	googleDocsAPIBase           = "https://docs.googleapis.com/v1"
	googleSheetsAPIBase         = "https://sheets.googleapis.com/v4"

	googleDocumentMimeType     = "application/vnd.google-apps.document"
	googleSpreadsheetMimeType  = "application/vnd.google-apps.spreadsheet"
	googlePresentationMimeType = "application/vnd.google-apps.presentation"
	maxProviderErrorBytes      = 8 << 10
	maxProviderJSONBytes       = 2 << 20
	maxThumbnailRedirects      = 3
)

type googleClient struct {
	http      *http.Client
	config    Config
	scopes    []string
	tokenURL  string
	driveURL  string
	docsURL   string
	sheetsURL string
}

type APIError struct {
	StatusCode int
	Code       string
	Message    string
	Reasons    []string
}

func (err *APIError) Error() string {
	if err == nil {
		return "Google API request failed"
	}
	if err.IsRateLimited() {
		return "Google rate limit reached; try again later"
	}
	switch err.StatusCode {
	case http.StatusUnauthorized:
		return "Google authorization expired"
	case http.StatusForbidden:
		return "Google denied access to this file"
	case http.StatusNotFound:
		return "Google file was not found"
	default:
		return fmt.Sprintf("Google API returned %d", err.StatusCode)
	}
}

func (err *APIError) IsRateLimited() bool {
	if err == nil {
		return false
	}
	if err.StatusCode == http.StatusTooManyRequests || err.Code == "resource_exhausted" {
		return true
	}
	for _, reason := range err.Reasons {
		switch strings.ToLower(strings.TrimSpace(reason)) {
		case "dailylimitexceeded", "ratelimitexceeded", "sharingratelimitexceeded", "userratelimitexceeded":
			return true
		}
	}
	return false
}

func (err *APIError) isFilePermissionLoss() bool {
	if err == nil {
		return false
	}
	for _, reason := range err.Reasons {
		switch strings.ToLower(strings.TrimSpace(reason)) {
		case "appnotauthorizedtofile", "insufficientfilepermissions":
			return true
		}
	}
	return false
}

func newGoogleClient(httpClient *http.Client, config Config, scopes []string) *googleClient {
	return &googleClient{
		http: httpClient, config: config, scopes: append([]string(nil), scopes...),
		tokenURL: googleTokenEndpoint, driveURL: googleDriveAPIBase,
		docsURL: googleDocsAPIBase, sheetsURL: googleSheetsAPIBase,
	}
}

func (client *googleClient) AuthorizationURL(state, verifier string) (string, error) {
	if client == nil || strings.TrimSpace(client.config.ClientID) == "" || strings.TrimSpace(client.config.RedirectURL) == "" {
		return "", errors.New("Google Drive OAuth is not configured")
	}
	challenge := sha256.Sum256([]byte(verifier))
	query := url.Values{
		"client_id": {client.config.ClientID}, "redirect_uri": {client.config.RedirectURL},
		"response_type": {"code"}, "scope": {strings.Join(client.scopes, " ")},
		"state": {state}, "access_type": {"offline"}, "prompt": {"consent select_account"},
		"code_challenge":        {base64.RawURLEncoding.EncodeToString(challenge[:])},
		"code_challenge_method": {"S256"},
	}
	return googleAuthorizationEndpoint + "?" + query.Encode(), nil
}

func (client *googleClient) Exchange(ctx context.Context, code, verifier string) (domain.OAuthToken, []string, error) {
	values := url.Values{
		"client_id": {client.config.ClientID}, "client_secret": {client.config.ClientSecret},
		"code": {code}, "code_verifier": {verifier}, "grant_type": {"authorization_code"},
		"redirect_uri": {client.config.RedirectURL},
	}
	return client.tokenRequest(ctx, values)
}

func (client *googleClient) Refresh(ctx context.Context, refreshToken string) (domain.OAuthToken, error) {
	values := url.Values{
		"client_id": {client.config.ClientID}, "client_secret": {client.config.ClientSecret},
		"refresh_token": {refreshToken}, "grant_type": {"refresh_token"},
	}
	token, _, err := client.tokenRequest(ctx, values)
	token.RefreshToken = refreshToken
	return token, err
}

func (client *googleClient) tokenRequest(ctx context.Context, values url.Values) (domain.OAuthToken, []string, error) {
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, client.tokenURL, strings.NewReader(values.Encode()))
	if err != nil {
		return domain.OAuthToken{}, nil, err
	}
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	response, err := client.http.Do(request)
	if err != nil {
		return domain.OAuthToken{}, nil, err
	}
	defer response.Body.Close()
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return domain.OAuthToken{}, nil, parseAPIError(response)
	}
	var body struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
		TokenType    string `json:"token_type"`
		ExpiresIn    int64  `json:"expires_in"`
		Scope        string `json:"scope"`
	}
	if err := decodeBoundedJSON(response.Body, &body); err != nil {
		return domain.OAuthToken{}, nil, err
	}
	if strings.TrimSpace(body.AccessToken) == "" {
		return domain.OAuthToken{}, nil, errors.New("Google returned an empty access token")
	}
	if body.ExpiresIn <= 0 {
		body.ExpiresIn = 3600
	}
	return domain.OAuthToken{
		AccessToken: body.AccessToken, RefreshToken: body.RefreshToken,
		TokenType: body.TokenType, Expiry: time.Now().UTC().Add(time.Duration(body.ExpiresIn) * time.Second),
	}, strings.Fields(body.Scope), nil
}

func (client *googleClient) UserInfo(ctx context.Context, accessToken string) (ProviderUser, error) {
	request, err := client.authenticatedRequest(ctx, http.MethodGet, googleUserInfoEndpoint, accessToken, nil)
	if err != nil {
		return ProviderUser{}, err
	}
	var response struct {
		Subject string `json:"sub"`
		Email   string `json:"email"`
		Name    string `json:"name"`
	}
	if err := client.doJSON(request, &response); err != nil {
		return ProviderUser{}, err
	}
	if strings.TrimSpace(response.Subject) == "" || strings.TrimSpace(response.Email) == "" {
		return ProviderUser{}, errors.New("Google account identity is incomplete")
	}
	return ProviderUser{Subject: response.Subject, Email: response.Email, DisplayName: optionalString(response.Name)}, nil
}

func (client *googleClient) Revoke(ctx context.Context, accessToken string) error {
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, googleRevokeEndpoint, strings.NewReader(url.Values{"token": {accessToken}}.Encode()))
	if err != nil {
		return err
	}
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	response, err := client.http.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	if response.StatusCode >= 200 && response.StatusCode < 300 {
		return nil
	}
	return parseAPIError(response)
}

func (client *googleClient) GetFile(ctx context.Context, accessToken, fileID string, resourceKey *string) (domain.ProviderFile, error) {
	query := url.Values{
		"fields":            {"id,name,mimeType,webViewLink,thumbnailLink,resourceKey,driveId,version,size,modifiedTime,trashed"},
		"supportsAllDrives": {"true"},
	}
	endpoint := client.driveURL + "/files/" + url.PathEscape(fileID) + "?" + query.Encode()
	request, err := client.authenticatedRequest(ctx, http.MethodGet, endpoint, accessToken, nil)
	if err != nil {
		return domain.ProviderFile{}, err
	}
	if err := setResourceKeyHeader(request, fileID, resourceKey); err != nil {
		return domain.ProviderFile{}, err
	}
	var body driveFileResponse
	if err := client.doJSON(request, &body); err != nil {
		return domain.ProviderFile{}, err
	}
	return body.providerFile()
}

func (client *googleClient) CreateFile(ctx context.Context, accessToken string, fileType domain.FileType, title, operationID string) (domain.ProviderFile, error) {
	mimeType := googleDocumentMimeType
	if fileType == domain.FileTypeSpreadsheet {
		mimeType = googleSpreadsheetMimeType
	}
	payload := map[string]any{
		"name": title, "mimeType": mimeType,
		"appProperties": map[string]string{"fortyoneOperationId": operationID},
	}
	encoded, err := json.Marshal(payload)
	if err != nil {
		return domain.ProviderFile{}, err
	}
	query := url.Values{"fields": {"id,name,mimeType,webViewLink,resourceKey,driveId,version,size,modifiedTime,trashed"}}
	request, err := client.authenticatedRequest(ctx, http.MethodPost, client.driveURL+"/files?"+query.Encode(), accessToken, bytes.NewReader(encoded))
	if err != nil {
		return domain.ProviderFile{}, err
	}
	request.Header.Set("Content-Type", "application/json")
	var body driveFileResponse
	if err := client.doJSON(request, &body); err != nil {
		return domain.ProviderFile{}, err
	}
	return body.providerFile()
}

func (client *googleClient) FindCreatedFile(ctx context.Context, accessToken, operationID string) (*domain.ProviderFile, error) {
	if _, err := uuidLike(operationID); err != nil {
		return nil, err
	}
	query := url.Values{
		"q":      {fmt.Sprintf("appProperties has { key='fortyoneOperationId' and value='%s' } and trashed=false", operationID)},
		"spaces": {"drive"}, "pageSize": {"2"},
		"fields": {"files(id,name,mimeType,webViewLink,resourceKey,driveId,version,size,modifiedTime,trashed)"},
	}
	request, err := client.authenticatedRequest(ctx, http.MethodGet, client.driveURL+"/files?"+query.Encode(), accessToken, nil)
	if err != nil {
		return nil, err
	}
	var response struct {
		Files []driveFileResponse `json:"files"`
	}
	if err := client.doJSON(request, &response); err != nil {
		return nil, err
	}
	if len(response.Files) == 0 {
		return nil, nil
	}
	file, err := response.Files[0].providerFile()
	return &file, err
}

func (client *googleClient) PopulateFile(ctx context.Context, accessToken string, file domain.ProviderFile, sourceURL string) error {
	if strings.TrimSpace(sourceURL) == "" {
		return nil
	}
	switch file.MimeType {
	case googleDocumentMimeType:
		payload := map[string]any{"requests": []map[string]any{{"insertText": map[string]any{
			"location": map[string]int{"index": 1}, "text": "Created from FortyOne\n" + sourceURL + "\n",
		}}}}
		return client.postJSON(ctx, accessToken, client.docsURL+"/documents/"+url.PathEscape(file.ID)+":batchUpdate", payload)
	case googleSpreadsheetMimeType:
		payload := map[string]any{"range": "A1:A2", "majorDimension": "ROWS", "values": [][]string{{"Created from FortyOne"}, {sourceURL}}}
		endpoint := client.sheetsURL + "/spreadsheets/" + url.PathEscape(file.ID) + "/values/A1:A2?valueInputOption=RAW"
		return client.putJSON(ctx, accessToken, endpoint, payload)
	default:
		return nil
	}
}

func (client *googleClient) ReadFile(ctx context.Context, accessToken string, file domain.ProviderFile, limit int64) (ProviderContent, error) {
	if limit <= 0 || limit > maxContentBytes {
		return ProviderContent{}, domain.ErrInvalidInput
	}
	contentType := "text/plain"
	var endpoint string
	switch file.MimeType {
	case googleDocumentMimeType, googlePresentationMimeType:
		endpoint = client.driveURL + "/files/" + url.PathEscape(file.ID) + "/export?" + url.Values{"mimeType": {"text/plain"}}.Encode()
	case googleSpreadsheetMimeType:
		contentType = "text/csv"
		endpoint = client.driveURL + "/files/" + url.PathEscape(file.ID) + "/export?" + url.Values{"mimeType": {contentType}}.Encode()
	case "text/plain", "text/csv":
		contentType = file.MimeType
		endpoint = client.driveURL + "/files/" + url.PathEscape(file.ID) + "?alt=media&supportsAllDrives=true"
	default:
		return ProviderContent{}, errors.New("this Google file type does not provide bounded text content")
	}
	request, err := client.authenticatedRequest(ctx, http.MethodGet, endpoint, accessToken, nil)
	if err != nil {
		return ProviderContent{}, err
	}
	if err := setResourceKeyHeader(request, file.ID, file.ResourceKey); err != nil {
		return ProviderContent{}, err
	}
	response, err := client.http.Do(request)
	if err != nil {
		return ProviderContent{}, err
	}
	defer response.Body.Close()
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return ProviderContent{}, parseAPIError(response)
	}
	data, err := io.ReadAll(io.LimitReader(response.Body, limit+1))
	if err != nil {
		return ProviderContent{}, err
	}
	truncated := int64(len(data)) > limit
	if truncated {
		data = data[:limit]
		// A byte limit can split the final UTF-8 code point. Remove only that
		// incomplete suffix; invalid bytes elsewhere still fail closed below.
		for removed := 0; removed < utf8.UTFMax-1 && !utf8.Valid(data) && len(data) > 0; removed++ {
			data = data[:len(data)-1]
		}
	}
	if !utf8.Valid(data) {
		return ProviderContent{}, errors.New("Google returned non-text content")
	}
	return ProviderContent{Text: string(data), ContentType: contentType, Truncated: truncated, BytesRead: len(data)}, nil
}

func (client *googleClient) ReadThumbnail(ctx context.Context, accessToken, thumbnailLink string, limit int64) (Preview, error) {
	if client == nil || client.http == nil || limit <= 0 || limit > maxPreviewBytes {
		return Preview{}, domain.ErrInvalidInput
	}
	thumbnailURL, err := validatedThumbnailURL(thumbnailLink)
	if err != nil {
		return Preview{}, err
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, thumbnailURL.String(), nil)
	if err != nil {
		return Preview{}, err
	}
	setThumbnailRequestHeaders(request, accessToken)

	httpClient := *client.http
	httpClient.Jar = nil
	configuredRedirect := client.http.CheckRedirect
	httpClient.CheckRedirect = func(redirect *http.Request, via []*http.Request) error {
		if len(via) > maxThumbnailRedirects {
			return errors.New("Google thumbnail exceeded the redirect limit")
		}
		if configuredRedirect != nil {
			if err := configuredRedirect(redirect, via); err != nil {
				return err
			}
		}
		if _, err := validatedThumbnailURL(redirect.URL.String()); err != nil {
			return fmt.Errorf("refuse unsafe Google thumbnail redirect: %w", err)
		}
		setThumbnailRequestHeaders(redirect, accessToken)
		return nil
	}

	response, err := httpClient.Do(request)
	if err != nil {
		var requestError *url.Error
		if errors.As(err, &requestError) && requestError.Err != nil {
			// thumbnailLink carries a short-lived provider capability in its query;
			// do not let net/url echo the complete URL into upstream logs.
			return Preview{}, requestError.Err
		}
		return Preview{}, err
	}
	defer response.Body.Close()
	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		return Preview{}, parseAPIError(response)
	}
	if response.ContentLength > limit {
		return Preview{}, domain.ErrContentTooLarge
	}
	declaredContentType, err := normalizedThumbnailContentType(response.Header.Get("Content-Type"))
	if err != nil {
		return Preview{}, err
	}
	data, err := io.ReadAll(io.LimitReader(response.Body, limit+1))
	if err != nil {
		return Preview{}, err
	}
	if int64(len(data)) > limit {
		return Preview{}, domain.ErrContentTooLarge
	}
	if len(data) == 0 {
		return Preview{}, errors.New("Google returned an empty thumbnail")
	}
	detectedContentType, err := normalizedThumbnailContentType(http.DetectContentType(data))
	if err != nil || detectedContentType != declaredContentType {
		return Preview{}, errors.New("Google returned invalid thumbnail content")
	}
	return Preview{Bytes: data, ContentType: detectedContentType}, nil
}

type driveFileResponse struct {
	ID            string `json:"id"`
	ResourceKey   string `json:"resourceKey"`
	Name          string `json:"name"`
	MimeType      string `json:"mimeType"`
	WebViewLink   string `json:"webViewLink"`
	ThumbnailLink string `json:"thumbnailLink"`
	DriveID       string `json:"driveId"`
	Version       string `json:"version"`
	Size          string `json:"size"`
	ModifiedAt    string `json:"modifiedTime"`
	Trashed       bool   `json:"trashed"`
}

func (file driveFileResponse) providerFile() (domain.ProviderFile, error) {
	if !validProviderFileID(file.ID) || strings.TrimSpace(file.Name) == "" || strings.TrimSpace(file.MimeType) == "" {
		return domain.ProviderFile{}, errors.New("Google returned incomplete file metadata")
	}
	modifiedAt, err := optionalTime(file.ModifiedAt)
	if err != nil {
		return domain.ProviderFile{}, err
	}
	var size *int64
	if file.Size != "" {
		parsed, parseErr := strconv.ParseInt(file.Size, 10, 64)
		if parseErr != nil || parsed < 0 {
			return domain.ProviderFile{}, errors.New("Google returned an invalid file size")
		}
		size = &parsed
	}
	webViewLink := strings.TrimSpace(file.WebViewLink)
	if webViewLink == "" {
		webViewLink = "https://drive.google.com/open?id=" + url.QueryEscape(file.ID)
	}
	metadata, _ := json.Marshal(map[string]bool{"trashed": file.Trashed})
	return domain.ProviderFile{
		ID: file.ID, ResourceKey: optionalString(file.ResourceKey), Name: file.Name,
		MimeType: file.MimeType, WebViewLink: webViewLink, ThumbnailLink: strings.TrimSpace(file.ThumbnailLink),
		DriveID: optionalString(file.DriveID),
		Version: optionalString(file.Version), SizeBytes: size, ModifiedAt: modifiedAt,
		Trashed: file.Trashed, Metadata: metadata,
	}, nil
}

func validatedThumbnailURL(rawURL string) (*url.URL, error) {
	parsed, err := url.Parse(strings.TrimSpace(rawURL))
	if err != nil || parsed == nil || !strings.EqualFold(parsed.Scheme, "https") || parsed.Opaque != "" ||
		parsed.User != nil || parsed.Hostname() == "" || parsed.Port() != "" || parsed.Fragment != "" ||
		!googleThumbnailHost(parsed.Hostname()) {
		return nil, errors.New("Google returned an unsafe thumbnail URL")
	}
	return parsed, nil
}

func googleThumbnailHost(host string) bool {
	host = strings.ToLower(strings.TrimSpace(host))
	return host == "drive.google.com" || host == "googleusercontent.com" || strings.HasSuffix(host, ".googleusercontent.com")
}

func setThumbnailRequestHeaders(request *http.Request, accessToken string) {
	request.Header.Set("Authorization", "Bearer "+accessToken)
	request.Header.Set("Accept", "image/webp,image/png,image/jpeg")
}

func normalizedThumbnailContentType(value string) (string, error) {
	contentType, _, err := mime.ParseMediaType(strings.TrimSpace(value))
	if err != nil {
		return "", errors.New("Google returned an invalid thumbnail content type")
	}
	contentType = strings.ToLower(contentType)
	switch contentType {
	case "image/jpeg", "image/png", "image/webp":
		return contentType, nil
	default:
		return "", errors.New("Google returned unsupported thumbnail content")
	}
}

func (client *googleClient) authenticatedRequest(ctx context.Context, method, endpoint, accessToken string, body io.Reader) (*http.Request, error) {
	request, err := http.NewRequestWithContext(ctx, method, endpoint, body)
	if err != nil {
		return nil, err
	}
	request.Header.Set("Authorization", "Bearer "+accessToken)
	request.Header.Set("Accept", "application/json")
	return request, nil
}

func (client *googleClient) doJSON(request *http.Request, target any) error {
	response, err := client.http.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return parseAPIError(response)
	}
	return decodeBoundedJSON(response.Body, target)
}

func (client *googleClient) postJSON(ctx context.Context, token, endpoint string, payload any) error {
	return client.writeJSON(ctx, http.MethodPost, token, endpoint, payload)
}

func (client *googleClient) putJSON(ctx context.Context, token, endpoint string, payload any) error {
	return client.writeJSON(ctx, http.MethodPut, token, endpoint, payload)
}

func (client *googleClient) writeJSON(ctx context.Context, method, token, endpoint string, payload any) error {
	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	request, err := client.authenticatedRequest(ctx, method, endpoint, token, bytes.NewReader(data))
	if err != nil {
		return err
	}
	request.Header.Set("Content-Type", "application/json")
	response, err := client.http.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return parseAPIError(response)
	}
	_, _ = io.Copy(io.Discard, io.LimitReader(response.Body, maxProviderJSONBytes))
	return nil
}

func decodeBoundedJSON(reader io.Reader, target any) error {
	decoder := json.NewDecoder(io.LimitReader(reader, maxProviderJSONBytes))
	if err := decoder.Decode(target); err != nil {
		return fmt.Errorf("decode Google response: %w", err)
	}
	return nil
}

func parseAPIError(response *http.Response) error {
	data, _ := io.ReadAll(io.LimitReader(response.Body, maxProviderErrorBytes))
	var body struct {
		Error json.RawMessage `json:"error"`
	}
	_ = json.Unmarshal(data, &body)
	code := ""
	reasons := make([]string, 0)
	if len(body.Error) > 0 {
		var oauthCode string
		if err := json.Unmarshal(body.Error, &oauthCode); err == nil {
			code = oauthCode
		} else {
			var providerError struct {
				Status string `json:"status"`
				Errors []struct {
					Reason string `json:"reason"`
				} `json:"errors"`
			}
			if err := json.Unmarshal(body.Error, &providerError); err == nil {
				code = providerError.Status
				for _, detail := range providerError.Errors {
					if reason := strings.TrimSpace(detail.Reason); reason != "" {
						reasons = append(reasons, reason)
					}
				}
			}
		}
	}
	return &APIError{StatusCode: response.StatusCode, Code: strings.ToLower(code), Reasons: reasons}
}

func isReauthorizationError(err error) bool {
	var apiError *APIError
	return errors.As(err, &apiError) && (apiError.StatusCode == http.StatusUnauthorized || apiError.Code == "invalid_grant")
}

func optionalString(value string) *string {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil
	}
	return &value
}

func optionalTime(value string) (*time.Time, error) {
	if strings.TrimSpace(value) == "" {
		return nil, nil
	}
	parsed, err := time.Parse(time.RFC3339Nano, value)
	if err != nil {
		return nil, fmt.Errorf("parse Google file modified time: %w", err)
	}
	return &parsed, nil
}

func setResourceKeyHeader(request *http.Request, fileID string, resourceKey *string) error {
	if request == nil {
		return domain.ErrInvalidInput
	}
	normalized, err := normalizeResourceKey(resourceKey)
	if err != nil {
		return err
	}
	if normalized == nil {
		return nil
	}
	request.Header.Set("X-Goog-Drive-Resource-Keys", fileID+"/"+*normalized)
	return nil
}

func uuidLike(value string) (string, error) {
	for _, character := range value {
		if (character >= 'a' && character <= 'f') || (character >= 'A' && character <= 'F') ||
			(character >= '0' && character <= '9') || character == '-' {
			continue
		}
		return "", domain.ErrInvalidInput
	}
	return value, nil
}
