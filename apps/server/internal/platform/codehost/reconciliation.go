package codehost

import (
	"crypto/sha256"
	"encoding/hex"
	"strings"
)

type SyncOrigin string

const SyncOriginFortyOne SyncOrigin = "fortyone"

// WorkItemSnapshot is the provider-neutral content used to detect whether an
// inbound work-item event merely echoes the last outbound write.
type WorkItemSnapshot struct {
	Title string
	Body  string
	State WorkItemState
}

type ReconciliationState struct {
	LastOrigin   SyncOrigin
	LastRevision string
}

// WorkItemRevision is stable across provider newline conventions. It is an
// echo-suppression fingerprint, not a security signature.
func WorkItemRevision(snapshot WorkItemSnapshot) string {
	normalizedBody := strings.ReplaceAll(snapshot.Body, "\r\n", "\n")
	normalizedBody = strings.ReplaceAll(normalizedBody, "\r", "\n")
	sum := sha256.Sum256([]byte(strings.Join([]string{
		snapshot.Title,
		normalizedBody,
		string(snapshot.State),
	}, "\x00")))
	return hex.EncodeToString(sum[:])
}

func ShouldSuppressEcho(state ReconciliationState, incomingRevision string) bool {
	return state.LastOrigin == SyncOriginFortyOne &&
		state.LastRevision != "" &&
		state.LastRevision == incomingRevision
}
