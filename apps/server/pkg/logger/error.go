package logger

import (
	"fmt"
	"io"
	"log/slog"
	"reflect"
	"strconv"
	"strings"
	"unicode"
)

const (
	maxDiagnosticCodeLength    = 96
	maxDiagnosticMessageLength = 160
	maxErrorChainNodes         = 16
	maxErrorChainTypes         = 8
)

// ErrorDefinition describes an operational error using bounded, reviewed
// metadata that is safe to write to production logs. Definitions should be
// declared as package-level values with literal arguments.
//
// The safe message must never contain err.Error(), provider response text,
// database statements, user input, credentials, or other runtime values.
type ErrorDefinition struct {
	code        string
	safeMessage string
}

// MustDefineError constructs an immutable production-safe error definition.
// It panics when a definition is malformed so invalid observability metadata is
// caught during process initialization instead of entering durable logs.
func MustDefineError(code, safeMessage string) ErrorDefinition {
	if !validDiagnosticCode(code) {
		panic("logger: invalid diagnostic error code")
	}
	if !validDiagnosticMessage(safeMessage) {
		panic("logger: invalid diagnostic safe message")
	}
	return ErrorDefinition{code: code, safeMessage: safeMessage}
}

// Wrap associates an underlying error with this definition while preserving
// errors.Is/errors.As traversal. The wrapper's own textual representation is
// deliberately limited to the reviewed safe message.
func (definition ErrorDefinition) Wrap(err error) error {
	if err == nil {
		return nil
	}
	if definition.code == "" || definition.safeMessage == "" {
		panic("logger: cannot use a zero error definition")
	}
	return &diagnosticError{definition: definition, cause: err}
}

// WrapIfUnclassified applies this definition only when the error chain does not
// already contain a more specific package-owned diagnostic. It is intended for
// process boundaries that need a fallback code without hiding a dependency or
// operation-level classification.
func (definition ErrorDefinition) WrapIfUnclassified(err error) error {
	if err == nil {
		return nil
	}
	if inspectError(err).code != "" {
		return err
	}
	return definition.Wrap(err)
}

type diagnosticError struct {
	definition ErrorDefinition
	cause      error
}

func (err *diagnosticError) Error() string {
	if err == nil || err.definition.safeMessage == "" {
		return "classified error"
	}
	return err.definition.safeMessage
}

func (err *diagnosticError) Unwrap() error {
	if err == nil {
		return nil
	}
	return err.cause
}

// Format prevents alternate fmt verbs such as %+v and %#v from exposing the
// wrapped cause through a structural representation.
func (err *diagnosticError) Format(state fmt.State, verb rune) {
	message := err.Error()
	if verb == 'q' {
		_, _ = io.WriteString(state, strconv.Quote(message))
		return
	}
	_, _ = io.WriteString(state, message)
}

func (err *diagnosticError) diagnosticMetadata() ErrorDefinition {
	if err == nil {
		return ErrorDefinition{}
	}
	return err.definition
}

type diagnosticMetadataCarrier interface {
	diagnosticMetadata() ErrorDefinition
}

type errorLogDetails struct {
	primaryType string
	typeChain   []string
	code        string
	safeMessage string
}

func safeErrorAttribute(key string, err error) slog.Attr {
	details := inspectError(err)
	attributes := []slog.Attr{
		slog.String("type", details.primaryType),
		slog.Any("type_chain", details.typeChain),
	}
	if details.code != "" {
		attributes = append(
			attributes,
			slog.String("code", details.code),
			slog.String("safe_message", details.safeMessage),
		)
	}
	return slog.Attr{Key: key, Value: slog.GroupValue(attributes...)}
}

// inspectError intentionally never calls Error on an arbitrary error. Error
// messages can contain credentials, provider payloads, SQL, URLs, or user data.
func inspectError(err error) (details errorLogDetails) {
	details.primaryType = redactedValue
	details.typeChain = []string{}
	defer func() {
		if recover() != nil {
			details = errorLogDetails{
				primaryType: redactedValue,
				typeChain:   []string{},
			}
		}
	}()

	visited := 0
	var visit func(error)
	visit = func(current error) {
		if current == nil || visited >= maxErrorChainNodes {
			return
		}
		visited++
		currentType := reflect.TypeOf(current)
		if currentType == nil {
			return
		}

		isNil := errorValueIsNil(current)
		if carrier, ok := current.(diagnosticMetadataCarrier); ok && !isNil {
			metadata := carrier.diagnosticMetadata()
			if details.code == "" && validDiagnosticCode(metadata.code) && validDiagnosticMessage(metadata.safeMessage) {
				details.code = metadata.code
				details.safeMessage = metadata.safeMessage
			}
		} else if len(details.typeChain) < maxErrorChainTypes {
			typeName := currentType.String()
			details.typeChain = append(details.typeChain, typeName)
			if details.primaryType == redactedValue {
				details.primaryType = typeName
			}
		}

		if isNil {
			return
		}
		for _, child := range unwrapErrors(current) {
			visit(child)
			if visited >= maxErrorChainNodes {
				return
			}
		}
	}
	visit(err)

	return details
}

func unwrapErrors(err error) (children []error) {
	defer func() {
		if recover() != nil {
			children = nil
		}
	}()
	if joined, ok := err.(interface{ Unwrap() []error }); ok {
		return joined.Unwrap()
	}
	if wrapped, ok := err.(interface{ Unwrap() error }); ok {
		if child := wrapped.Unwrap(); child != nil {
			return []error{child}
		}
	}
	return nil
}

func errorValueIsNil(err error) bool {
	value := reflect.ValueOf(err)
	switch value.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Pointer, reflect.Slice:
		return value.IsNil()
	default:
		return false
	}
}

func validDiagnosticCode(code string) bool {
	if code == "" || len(code) > maxDiagnosticCodeLength {
		return false
	}
	segmentStart := true
	for _, character := range code {
		switch {
		case character == '.':
			if segmentStart {
				return false
			}
			segmentStart = true
		case character >= 'a' && character <= 'z':
			segmentStart = false
		case !segmentStart && ((character >= '0' && character <= '9') || character == '_'):
		default:
			return false
		}
	}
	return !segmentStart
}

func validDiagnosticMessage(message string) bool {
	if message == "" || len(message) > maxDiagnosticMessageLength || strings.TrimSpace(message) != message {
		return false
	}
	for _, character := range message {
		if unicode.IsControl(character) {
			return false
		}
	}
	return true
}
