package web

import (
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/google/uuid"
)

const DefaultMaxQueryParameterBytes = 4096

const maxOAuthProviderErrorBytes = 256

var (
	ErrInvalidQueryParameter  = errors.New("query parameter is invalid")
	ErrRepeatedQueryParameter = errors.New("query parameter must appear exactly once")
	ErrQueryParameterTooLong  = errors.New("query parameter is too long")
)

// QueryParameterError deliberately excludes the supplied value so bearer-like
// or customer-controlled query data cannot be copied into logs by generic
// error handling.
type QueryParameterError struct {
	Name  string
	Cause error
}

// QueryListLimits bounds every dimension of a list-valued query parameter.
// MaxBytes applies to the combined encoded values before splitting, MaxItemBytes
// applies after trimming, and MaxItems applies after comma/repeated-value
// expansion.
type QueryListLimits struct {
	MaxBytes     int
	MaxItemBytes int
	MaxItems     int
}

// OAuthCallbackQuery is the bounded, unambiguous representation of the
// standard OAuth authorization-code callback parameters. ProviderError is
// deliberately opaque; handlers should map it to a small internal error code
// before logging or returning it.
type OAuthCallbackQuery struct {
	Code          string
	State         string
	ProviderError string
}

func (e *QueryParameterError) Error() string {
	return fmt.Sprintf("%s: %v", e.Name, e.Cause)
}

func (e *QueryParameterError) Unwrap() error {
	return e.Cause
}

// OptionalQueryParameter reads one bounded value. Repeated parameters are
// rejected instead of silently selecting one attacker-controlled value.
func OptionalQueryParameter(values url.Values, name string, maxBytes int) (string, bool, error) {
	name = strings.TrimSpace(name)
	if name == "" || maxBytes <= 0 {
		return "", false, &QueryParameterError{Name: "query parameter", Cause: ErrInvalidQueryParameter}
	}
	raw, exists := values[name]
	if !exists {
		return "", false, nil
	}
	if len(raw) != 1 {
		return "", false, &QueryParameterError{Name: name, Cause: ErrRepeatedQueryParameter}
	}
	if len(raw[0]) > maxBytes {
		return "", false, &QueryParameterError{Name: name, Cause: ErrQueryParameterTooLong}
	}
	return raw[0], true, nil
}

// OptionalOpaqueQueryParameter reads one exact protocol value while rejecting
// invalid UTF-8, NULs, and control characters. Unlike the text helper it does
// not trim or otherwise rewrite the value, which matters for signed tokens and
// provider-issued authorization codes.
func OptionalOpaqueQueryParameter(values url.Values, name string, maxBytes int) (string, bool, error) {
	value, present, err := OptionalQueryParameter(values, name, maxBytes)
	if err != nil || !present {
		return "", present, err
	}
	if !utf8.ValidString(value) || strings.IndexFunc(value, unicode.IsControl) >= 0 {
		return "", false, &QueryParameterError{Name: strings.TrimSpace(name), Cause: ErrInvalidQueryParameter}
	}
	return value, true, nil
}

// RequiredOpaqueQueryParameter is the required counterpart to
// OptionalOpaqueQueryParameter. The original bytes are returned unchanged,
// but a missing or whitespace-only protocol value is invalid.
func RequiredOpaqueQueryParameter(values url.Values, name string, maxBytes int) (string, error) {
	value, present, err := OptionalOpaqueQueryParameter(values, name, maxBytes)
	if err != nil {
		return "", err
	}
	if !present || strings.TrimSpace(value) == "" {
		return "", &QueryParameterError{Name: strings.TrimSpace(name), Cause: ErrInvalidQueryParameter}
	}
	return value, nil
}

// OptionalTextQueryParameter reads one bounded UTF-8 text value, rejects NUL
// bytes and excessive rune counts, and returns the trimmed value. The byte and
// rune limits are both explicit because storage/search contracts often bound
// characters while ingress must still cap encoded memory.
func OptionalTextQueryParameter(values url.Values, name string, maxBytes, maxRunes int) (string, bool, error) {
	if maxRunes <= 0 {
		return "", false, &QueryParameterError{Name: strings.TrimSpace(name), Cause: ErrInvalidQueryParameter}
	}
	value, present, err := OptionalQueryParameter(values, name, maxBytes)
	if err != nil || !present {
		return "", present, err
	}
	if !utf8.ValidString(value) || strings.ContainsRune(value, '\x00') || utf8.RuneCountInString(value) > maxRunes {
		return "", false, &QueryParameterError{Name: strings.TrimSpace(name), Cause: ErrInvalidQueryParameter}
	}
	return strings.TrimSpace(value), true, nil
}

// RequiredTextQueryParameter reads one non-blank bounded UTF-8 text value. It
// has the same ambiguity and value-leak protections as
// OptionalTextQueryParameter, but treats a missing or whitespace-only value as
// invalid. OAuth callback state/code parameters and other protocol-required
// text should use this helper instead of URL.Query().Get.
func RequiredTextQueryParameter(values url.Values, name string, maxBytes, maxRunes int) (string, error) {
	value, present, err := OptionalTextQueryParameter(values, name, maxBytes, maxRunes)
	if err != nil {
		return "", err
	}
	if !present || value == "" {
		return "", &QueryParameterError{Name: strings.TrimSpace(name), Cause: ErrInvalidQueryParameter}
	}
	return value, nil
}

// ParseOAuthCallbackQuery validates the standard code/state/error callback
// shape. State is always required. Code is required for a successful callback
// and optional when the provider supplied an error. Every present scalar is
// still validated so an error callback cannot hide a repeated or oversized
// code parameter.
func ParseOAuthCallbackQuery(values url.Values) (OAuthCallbackQuery, error) {
	state, err := RequiredOpaqueQueryParameter(
		values,
		"state",
		DefaultMaxQueryParameterBytes,
	)
	if err != nil {
		return OAuthCallbackQuery{}, err
	}
	code, codePresent, err := OptionalOpaqueQueryParameter(
		values,
		"code",
		DefaultMaxQueryParameterBytes,
	)
	if err != nil {
		return OAuthCallbackQuery{}, err
	}
	providerError, _, err := OptionalOpaqueQueryParameter(
		values,
		"error",
		maxOAuthProviderErrorBytes,
	)
	if err != nil {
		return OAuthCallbackQuery{}, err
	}
	if providerError == "" && (!codePresent || code == "") {
		return OAuthCallbackQuery{}, &QueryParameterError{Name: "code", Cause: ErrInvalidQueryParameter}
	}
	return OAuthCallbackQuery{Code: code, State: state, ProviderError: providerError}, nil
}

// OptionalListQueryParameter reads a bounded list whose callers may encode
// items as repeated parameters, comma-separated values, or both. Empty items
// are rejected and duplicates are removed while preserving first-seen order.
// This is intentionally separate from OptionalQueryParameter: repetition is
// unambiguous for a list but must remain an error for scalar parameters.
func OptionalListQueryParameter(values url.Values, name string, limits QueryListLimits) ([]string, bool, error) {
	name = strings.TrimSpace(name)
	if name == "" || limits.MaxBytes <= 0 || limits.MaxItemBytes <= 0 || limits.MaxItems <= 0 {
		return nil, false, &QueryParameterError{Name: "query parameter", Cause: ErrInvalidQueryParameter}
	}
	rawValues, exists := values[name]
	if !exists {
		return nil, false, nil
	}

	totalBytes := 0
	items := make([]string, 0, min(len(rawValues), limits.MaxItems))
	seen := make(map[string]struct{}, min(len(rawValues), limits.MaxItems))
	for _, raw := range rawValues {
		if len(raw) > limits.MaxBytes-totalBytes {
			return nil, false, &QueryParameterError{Name: name, Cause: ErrQueryParameterTooLong}
		}
		totalBytes += len(raw)
		for _, item := range strings.Split(raw, ",") {
			item = strings.TrimSpace(item)
			if item == "" {
				return nil, false, &QueryParameterError{Name: name, Cause: ErrInvalidQueryParameter}
			}
			if len(item) > limits.MaxItemBytes {
				return nil, false, &QueryParameterError{Name: name, Cause: ErrQueryParameterTooLong}
			}
			if _, duplicate := seen[item]; duplicate {
				continue
			}
			if len(items) == limits.MaxItems {
				return nil, false, &QueryParameterError{Name: name, Cause: ErrInvalidQueryParameter}
			}
			seen[item] = struct{}{}
			items = append(items, item)
		}
	}
	if len(items) == 0 {
		return nil, false, &QueryParameterError{Name: name, Cause: ErrInvalidQueryParameter}
	}
	return items, true, nil
}

// OptionalIntegerQueryParameter parses one bounded base-10 integer within an
// explicit inclusive range. It rejects blank, repeated, oversized, malformed,
// overflowing, and out-of-range values without echoing the supplied value.
func OptionalIntegerQueryParameter(
	values url.Values,
	name string,
	maxBytes int,
	minimum int,
	maximum int,
) (int, bool, error) {
	if minimum > maximum {
		return 0, false, &QueryParameterError{Name: strings.TrimSpace(name), Cause: ErrInvalidQueryParameter}
	}
	value, present, err := OptionalQueryParameter(values, name, maxBytes)
	if err != nil || !present {
		return 0, present, err
	}
	parsed, err := strconv.Atoi(strings.TrimSpace(value))
	if err != nil || parsed < minimum || parsed > maximum {
		return 0, false, &QueryParameterError{Name: strings.TrimSpace(name), Cause: ErrInvalidQueryParameter}
	}
	return parsed, true, nil
}

// OptionalBooleanQueryParameter parses one canonical boolean value. Only the
// lowercase protocol values "true" and "false" are accepted; accepting the
// wider strconv.ParseBool grammar would make the wire contract ambiguous.
func OptionalBooleanQueryParameter(values url.Values, name string) (bool, bool, error) {
	value, present, err := OptionalQueryParameter(values, name, DefaultMaxQueryParameterBytes)
	if err != nil || !present {
		return false, present, err
	}
	switch value {
	case "true":
		return true, true, nil
	case "false":
		return false, true, nil
	default:
		return false, false, &QueryParameterError{Name: strings.TrimSpace(name), Cause: ErrInvalidQueryParameter}
	}
}

func OptionalUUIDQueryParameter(values url.Values, name string) (*uuid.UUID, error) {
	value, present, err := OptionalQueryParameter(values, name, DefaultMaxQueryParameterBytes)
	if err != nil {
		return nil, err
	}
	value = strings.TrimSpace(value)
	if !present || value == "" {
		return nil, nil
	}
	parsed, err := uuid.Parse(value)
	if err != nil || parsed == uuid.Nil {
		return nil, &QueryParameterError{Name: strings.TrimSpace(name), Cause: ErrInvalidQueryParameter}
	}
	return &parsed, nil
}

// OptionalCommaSeparatedUUIDQueryParameter reads a bounded comma-separated
// UUID list. Empty segments are ignored for backwards compatibility, while a
// zero UUID, malformed UUID, repeated parameter, or excessive item count is
// rejected without including the supplied value in the error.
func OptionalCommaSeparatedUUIDQueryParameter(values url.Values, name string, maxItems int) ([]uuid.UUID, error) {
	if maxItems <= 0 {
		return nil, &QueryParameterError{Name: strings.TrimSpace(name), Cause: ErrInvalidQueryParameter}
	}
	value, present, err := OptionalQueryParameter(values, name, DefaultMaxQueryParameterBytes)
	if err != nil {
		return nil, err
	}
	if !present || strings.TrimSpace(value) == "" {
		return nil, nil
	}

	parts := strings.Split(value, ",")
	result := make([]uuid.UUID, 0, min(len(parts), maxItems))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		if len(result) == maxItems {
			return nil, &QueryParameterError{Name: strings.TrimSpace(name), Cause: ErrInvalidQueryParameter}
		}
		parsed, err := uuid.Parse(part)
		if err != nil || parsed == uuid.Nil {
			return nil, &QueryParameterError{Name: strings.TrimSpace(name), Cause: ErrInvalidQueryParameter}
		}
		result = append(result, parsed)
	}
	return result, nil
}
