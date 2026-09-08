package logger

import (
	"regexp"
	"strings"
	"unicode/utf8"
)

const maxErrorMessageBytes = 4096

var (
	errorPrivateKey    = regexp.MustCompile(`(?s)-----BEGIN [^-]*PRIVATE KEY-----.*?(?:-----END [^-]*PRIVATE KEY-----|$)`)
	errorURL           = regexp.MustCompile(`(?i)\b[a-z][a-z0-9+.-]*://[^\s<>"']+`)
	errorEmail         = regexp.MustCompile(`(?i)\b[a-z0-9.!#$%&'*+/=?^_` + "`" + `{|}~-]+@[a-z0-9.-]+\.[a-z]{2,}\b`)
	errorCredential    = regexp.MustCompile(`(?i)(["']?(?:authorization|proxy-authorization|cookie|set-cookie|[a-z0-9_.-]*(?:password|passwd|secret|token|api[_-]?key))["']?\s*[:=]\s*)(?:"[^"\n]*"|'[^'\n]*'|[^\s,;}]+)`)
	errorAuthorization = regexp.MustCompile(`(?i)\b(?:Bearer|Basic)\s+[a-z0-9._~+/=-]+`)
	errorToken         = regexp.MustCompile(`\b(?:eyJ[a-zA-Z0-9_-]+\.[a-zA-Z0-9_-]+\.[a-zA-Z0-9_-]+|(?:sk-|ghp_|github_pat_|xox[baprs]-)[a-zA-Z0-9_-]+)`)
	// Provider response bodies and PostgreSQL DETAIL can contain arbitrary user
	// data. Preserve the operation and error summary preceding those sections.
	errorPayload = regexp.MustCompile(`(?is)\b(?:provider\s+(?:response|body)(?:\s|[:=])|(?:response\s+body|DETAIL|STATEMENT|raw[_ ]payload)\s*[:=]).*`)
)

// Error text is useful operational evidence, but must not bypass the logger's
// credential and payload redaction. This is a fallback for unclassified errors;
// reviewed ErrorDefinitions remain the preferred diagnostic contract.
func readableErrorMessage(err error) (message string) {
	message = "error message unavailable"
	defer func() { _ = recover() }()
	if err == nil || errorValueIsNil(err) {
		return message
	}
	return redactErrorMessage(err.Error())
}

// Classified wrappers intentionally expose only a reviewed summary through
// Error(). Include their redacted causes so startup and worker diagnostics
// still explain which configuration or dependency failed.
func readableErrorCauses(err error) []string {
	seen := map[string]bool{readableErrorMessage(err): true}
	var messages []string
	visited, remaining := 0, maxErrorMessageBytes
	var visit func(error)
	visit = func(current error) {
		if current == nil || errorValueIsNil(current) || visited >= maxErrorChainNodes || remaining <= 0 {
			return
		}
		visited++
		message := readableErrorMessage(current)
		if !seen[message] {
			seen[message] = true
			if len(message) > remaining {
				return
			}
			remaining -= len(message)
			messages = append(messages, message)
		}
		for _, child := range unwrapErrors(current) {
			visit(child)
		}
	}
	visit(err)
	return messages
}

func redactErrorMessage(message string) string {
	message = strings.ToValidUTF8(message, "�")
	message = errorPrivateKey.ReplaceAllString(message, redactedValue)
	message = errorPayload.ReplaceAllString(message, redactedValue)
	message = errorURL.ReplaceAllString(message, redactedValue)
	message = errorEmail.ReplaceAllString(message, redactedValue)
	message = errorAuthorization.ReplaceAllString(message, redactedValue)
	message = errorCredential.ReplaceAllString(message, "${1}"+redactedValue)
	message = errorToken.ReplaceAllString(message, redactedValue)
	if len(message) > maxErrorMessageBytes {
		message = message[:maxErrorMessageBytes]
		for !utf8.ValidString(message) {
			message = message[:len(message)-1]
		}
		message += "… [truncated]"
	}
	return message
}
