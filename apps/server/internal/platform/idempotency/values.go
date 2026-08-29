// Package idempotency provides the shared, transport-neutral receipt lifecycle
// used to make explicitly adopted API mutations safe to retry.
package idempotency

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"io"
	"mime"
	"regexp"
	"strings"

	"github.com/complexus-tech/projects-api/internal/platform/auth"
	"github.com/google/uuid"
)

const (
	MinKeyBytes          = 16
	MaxKeyBytes          = 255
	MaxOperationBytes    = 128
	MaxRequestBodyBytes  = 1024 * 1024
	MaxResponseBodyBytes = 64 * 1024
	MaxContentTypeBytes  = 128
)

var (
	ErrInvalidKey       = errors.New("invalid idempotency key")
	ErrInvalidMethod    = errors.New("invalid idempotency HTTP method")
	ErrInvalidOperation = errors.New("invalid idempotency route operation")
	ErrInvalidScope     = errors.New("invalid idempotency scope")
	ErrRequestTooLarge  = errors.New("idempotency request body exceeds the supported limit")
	ErrInvalidResponse  = errors.New("invalid idempotency response")

	operationPattern = regexp.MustCompile(`^[a-z][a-z0-9._:-]{0,127}$`)
)

// Key is an exact, bounded caller-supplied idempotency key. It deliberately
// has no String or text-marshalling method so it is difficult to log or expose
// accidentally. Persistence receives only its SHA-256 digest.
type Key struct {
	value string
}

func ParseKey(value string) (Key, error) {
	if len(value) < MinKeyBytes || len(value) > MaxKeyBytes {
		return Key{}, fmt.Errorf(
			"%w: length must be between %d and %d bytes",
			ErrInvalidKey,
			MinKeyBytes,
			MaxKeyBytes,
		)
	}
	for index := 0; index < len(value); index++ {
		if value[index] < 0x21 || value[index] > 0x7e {
			return Key{}, fmt.Errorf("%w: only visible ASCII bytes are accepted", ErrInvalidKey)
		}
	}
	return Key{value: value}, nil
}

// Format keeps the raw key redacted even if a caller passes the value to a
// structured or printf-style logger by mistake.
func (Key) Format(state fmt.State, _ rune) {
	_, _ = io.WriteString(state, "[REDACTED]")
}

func (k Key) validate() error {
	_, err := ParseKey(k.value)
	return err
}

func (k Key) digest() [sha256.Size]byte {
	return sha256.Sum256([]byte(k.value))
}

// RequestHash is the SHA-256 digest of the exact request bytes covered by a
// receipt. Callers must pass the same canonical bytes that the handler uses.
type RequestHash [sha256.Size]byte

func HashRequest(body []byte) RequestHash {
	return sha256.Sum256(body)
}

func validateRequestBody(body []byte) error {
	if len(body) > MaxRequestBodyBytes {
		return fmt.Errorf("%w: maximum is %d bytes", ErrRequestTooLarge, MaxRequestBodyBytes)
	}
	return nil
}

type Method string

const (
	MethodPost   Method = "POST"
	MethodPut    Method = "PUT"
	MethodPatch  Method = "PATCH"
	MethodDelete Method = "DELETE"
)

func ParseMethod(value string) (Method, error) {
	method := Method(strings.ToUpper(value))
	switch method {
	case MethodPost, MethodPut, MethodPatch, MethodDelete:
		return method, nil
	default:
		return "", fmt.Errorf("%w: only mutation methods are supported", ErrInvalidMethod)
	}
}

func (m Method) validate() error {
	parsed, err := ParseMethod(string(m))
	if err != nil || parsed != m {
		return ErrInvalidMethod
	}
	return nil
}

// Operation is a stable route identifier such as "stories.create". It must
// never include a raw URL, path parameter, workspace slug, or resource ID.
type Operation struct {
	value string
}

func ParseOperation(value string) (Operation, error) {
	if len(value) == 0 || len(value) > MaxOperationBytes || !operationPattern.MatchString(value) {
		return Operation{}, fmt.Errorf(
			"%w: use 1-%d lowercase ASCII letters, digits, dot, underscore, colon, or hyphen",
			ErrInvalidOperation,
			MaxOperationBytes,
		)
	}
	return Operation{value: value}, nil
}

func (o Operation) validate() error {
	_, err := ParseOperation(o.value)
	return err
}

// Scope binds a key to the authenticated principal, its optional selected
// workspace, the mutation method, and a stable route operation. Workspace
// identity is derived from Actor so a caller cannot supply a mismatched tenant.
type Scope struct {
	principalKind auth.PrincipalKind
	principalID   uuid.UUID
	workspaceID   uuid.UUID
	method        Method
	operation     Operation
}

func NewScope(actor auth.Actor, method Method, operation Operation) (Scope, error) {
	if err := actor.Validate(); err != nil {
		return Scope{}, fmt.Errorf("%w: actor: %v", ErrInvalidScope, err)
	}
	identityID := actor.PrincipalID
	if actor.Kind == auth.PrincipalOAuthApplication {
		if actor.CredentialID == uuid.Nil {
			return Scope{}, fmt.Errorf("%w: OAuth application installation id is required", ErrInvalidScope)
		}
		// OAuth application principals describe who acted, while the
		// installation is the stable credential boundary for retries. Access
		// token IDs are intentionally excluded because they rotate on every
		// client-credentials exchange.
		identityID = actor.CredentialID
	}
	if err := method.validate(); err != nil {
		return Scope{}, fmt.Errorf("%w: method: %v", ErrInvalidScope, err)
	}
	if err := operation.validate(); err != nil {
		return Scope{}, fmt.Errorf("%w: operation: %v", ErrInvalidScope, err)
	}
	return Scope{
		principalKind: actor.Kind,
		principalID:   identityID,
		workspaceID:   actor.WorkspaceID,
		method:        method,
		operation:     operation,
	}, nil
}

func (s Scope) validate() error {
	if s.principalID == uuid.Nil {
		return fmt.Errorf("%w: principal id is required", ErrInvalidScope)
	}
	if err := s.method.validate(); err != nil {
		return fmt.Errorf("%w: method: %v", ErrInvalidScope, err)
	}
	if err := s.operation.validate(); err != nil {
		return fmt.Errorf("%w: operation: %v", ErrInvalidScope, err)
	}
	return nil
}

// Response contains the only replay metadata retained by the shared service.
// Arbitrary headers, including Set-Cookie and authentication headers, are not
// representable and therefore cannot be persisted or replayed.
type Response struct {
	statusCode  int
	body        []byte
	contentType string
}

func NewResponse(statusCode int, body []byte, contentType string) (Response, error) {
	if statusCode < 200 || statusCode > 599 {
		return Response{}, fmt.Errorf("%w: status code must be between 200 and 599", ErrInvalidResponse)
	}
	if len(body) > MaxResponseBodyBytes {
		return Response{}, fmt.Errorf("%w: body exceeds %d bytes", ErrInvalidResponse, MaxResponseBodyBytes)
	}
	if len(contentType) == 0 || len(contentType) > MaxContentTypeBytes || strings.ContainsAny(contentType, "\r\n") {
		return Response{}, fmt.Errorf("%w: content type is missing, too long, or contains a line break", ErrInvalidResponse)
	}
	if _, _, err := mime.ParseMediaType(contentType); err != nil {
		return Response{}, fmt.Errorf("%w: content type is malformed", ErrInvalidResponse)
	}
	return Response{
		statusCode:  statusCode,
		body:        cloneBytes(body),
		contentType: contentType,
	}, nil
}

func (r Response) validate() error {
	_, err := NewResponse(r.statusCode, r.body, r.contentType)
	return err
}

func (r Response) StatusCode() int {
	return r.statusCode
}

func (r Response) Body() []byte {
	return cloneBytes(r.body)
}

func (r Response) ContentType() string {
	return r.contentType
}

func cloneBytes(value []byte) []byte {
	cloned := make([]byte, len(value))
	copy(cloned, value)
	return cloned
}
