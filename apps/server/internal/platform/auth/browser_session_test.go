package auth

import (
	"errors"
	"testing"

	"github.com/google/uuid"
)

func TestBrowserSessionValidation(t *testing.T) {
	t.Parallel()

	valid, err := NewBrowserSession(uuid.New(), 1)
	if err != nil {
		t.Fatalf("new browser session: %v", err)
	}
	if err := valid.Validate(); err != nil {
		t.Fatalf("validate browser session: %v", err)
	}

	for _, session := range []BrowserSession{
		{UserID: uuid.Nil, Version: 1},
		{UserID: uuid.New(), Version: 0},
		{UserID: uuid.New(), Version: -1},
	} {
		if err := session.Validate(); !errors.Is(err, ErrInvalidBrowserSession) {
			t.Fatalf("Validate(%+v) error = %v, want ErrInvalidBrowserSession", session, err)
		}
	}
}
