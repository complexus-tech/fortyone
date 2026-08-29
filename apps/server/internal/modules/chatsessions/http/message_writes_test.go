package chatsessionshttp

import (
	"errors"
	"net/http"
	"os"
	"strings"
	"testing"

	chatsessions "github.com/complexus-tech/projects-api/internal/modules/chatsessions/service"
	"github.com/google/uuid"
)

func TestPublicRoutesDoNotExposeUntrustedMutationReconciliation(t *testing.T) {
	t.Parallel()

	source, err := os.ReadFile("routes.go")
	if err != nil {
		t.Fatalf("read chat session routes: %v", err)
	}
	if strings.Contains(string(source), "/reconcile") {
		t.Fatal("mutation reconciliation must remain internal until evidence is independently verified")
	}
}

func TestMessageWriteRequestStatus(t *testing.T) {
	t.Parallel()

	tests := []struct {
		err  error
		want int
	}{
		{err: chatsessions.ErrMessageWriteConflict, want: http.StatusConflict},
		{err: chatsessions.ErrMessageWriteApprovalOpen, want: http.StatusConflict},
		{err: chatsessions.ErrMessageWriteInvalid, want: http.StatusUnprocessableEntity},
		{err: chatsessions.ErrNotFound, want: http.StatusNotFound},
		{err: errors.New("repository unavailable"), want: http.StatusInternalServerError},
	}

	for _, test := range tests {
		if got := messageWriteRequestStatus(test.err); got != test.want {
			t.Errorf("messageWriteRequestStatus(%v) = %d, want %d", test.err, got, test.want)
		}
	}
}

func TestMessageWriteRequestValidation(t *testing.T) {
	t.Parallel()

	if err := (AppBeginMessageWriteRequest{
		Title:     "Chat",
		Operation: chatsessions.MessageWriteAppend,
	}).Validate(); err != nil {
		t.Fatalf("valid begin request: %v", err)
	}
	if err := (AppBeginMessageWriteRequest{
		Title:     "Chat",
		Operation: "rewrite",
	}).Validate(); err == nil {
		t.Fatal("unsupported operation must fail")
	}
	if err := (AppFinalizeMessageWriteRequest{
		Generation: 1,
		Token:      uuid.New(),
	}).Validate(); err != nil {
		t.Fatalf("valid finalize request: %v", err)
	}
	if err := (AppFinalizeMessageWriteRequest{}).Validate(); err == nil {
		t.Fatal("empty finalize reservation must fail")
	}
	if err := (AppRecoverMutationApprovalOutputRequest{
		Fingerprint: strings.Repeat("a", 64),
	}).Validate(); err != nil {
		t.Fatalf("valid recovery request: %v", err)
	}
}
