package chatservice

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	domain "github.com/complexus-tech/projects-api/internal/modules/chat/domain"
	"github.com/complexus-tech/projects-api/internal/platform/auth"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

type testRepository struct {
	Repository
	create func(domain.CreateMessageCommand) (domain.Message, error)
}

func (r testRepository) WithinUnit(ctx context.Context, operation func(context.Context) error) error {
	return operation(ctx)
}

func (r testRepository) CreateMessage(_ context.Context, c domain.CreateMessageCommand, _ uuid.UUID) (domain.Message, error) {
	return r.create(c)
}

func testScope(t *testing.T) domain.ActorScope {
	t.Helper()
	workspace := uuid.New()
	actor, err := auth.NewHumanActor(uuid.New()).WithWorkspace(workspace)
	require.NoError(t, err)
	return domain.ActorScope{WorkspaceID: workspace, Actor: actor}
}

func TestCreateValidatesBeforeWritingAndCallsHookOnlyOnce(t *testing.T) {
	scope := testScope(t)
	command := domain.CreateMessageCommand{Scope: scope, ConversationID: uuid.New(), CreationID: uuid.New(), Text: "Hello"}
	created := domain.Message{ID: uuid.New()}
	calls, hooks := 0, 0
	repo := testRepository{create: func(c domain.CreateMessageCommand) (domain.Message, error) {
		calls++
		require.Equal(t, command, c)
		return created, nil
	}}
	service := New(repo, WithMessageCreatedHook(func(_ context.Context, c domain.CreateMessageCommand, m domain.Message) error {
		hooks++
		require.Equal(t, command, c)
		require.Equal(t, created.ID, m.ID)
		return nil
	}))
	_, err := service.CreateMessage(t.Context(), command)
	require.NoError(t, err)
	require.Equal(t, 1, hooks)
	created.Replayed = true
	_, err = service.CreateMessage(t.Context(), command)
	require.NoError(t, err)
	require.Equal(t, 1, hooks)
	invalid := command
	invalid.Text = strings.Repeat("é", 4001)
	_, err = service.CreateMessage(t.Context(), invalid)
	require.ErrorIs(t, err, domain.ErrInvalid)
	require.Equal(t, 2, calls)
	invalid = command
	invalid.Scope.WorkspaceID = uuid.New()
	_, err = service.CreateMessage(t.Context(), invalid)
	require.ErrorIs(t, err, domain.ErrForbidden)
	require.Equal(t, 2, calls)
}

func TestContentValidationBoundsStructuredInput(t *testing.T) {
	id := uuid.New()
	tests := []struct {
		name, text            string
		rich                  json.RawMessage
		mentions, attachments []uuid.UUID
		agents                []string
		valid                 bool
	}{
		{name: "unicode limit", text: strings.Repeat("é", 4000), valid: true},
		{name: "file only", attachments: []uuid.UUID{id}, valid: true},
		{name: "empty", text: "   "},
		{name: "unknown agent", text: "hello", agents: []string{"other"}},
		{name: "duplicate mention", text: "hello", mentions: []uuid.UUID{id, id}},
		{name: "array document", text: "hello", rich: json.RawMessage(`[]`)},
		{name: "oversized document", text: "hello", rich: json.RawMessage(`{"text":"` + strings.Repeat("a", 65536) + `"}`)},
		{name: "deep document", text: "hello", rich: json.RawMessage(strings.Repeat(`{"content":`, 25) + `null` + strings.Repeat(`}`, 25))},
		{name: "bounded document", text: "hello", rich: json.RawMessage(`{"type":"doc","content":[{"type":"paragraph"}]}`), valid: true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := validateContent(test.text, test.rich, test.mentions, test.agents, test.attachments)
			if test.valid {
				require.NoError(t, err)
			} else {
				require.ErrorIs(t, err, domain.ErrInvalid)
			}
		})
	}
}

func TestHookFailurePropagatesToTransaction(t *testing.T) {
	failure := errors.New("invocation acceptance unavailable")
	service := New(testRepository{create: func(domain.CreateMessageCommand) (domain.Message, error) { return domain.Message{ID: uuid.New()}, nil }}, WithMessageCreatedHook(func(context.Context, domain.CreateMessageCommand, domain.Message) error { return failure }))
	_, err := service.CreateMessage(t.Context(), domain.CreateMessageCommand{Scope: testScope(t), ConversationID: uuid.New(), CreationID: uuid.New(), Text: "@Maya help"})
	require.ErrorIs(t, err, failure)
}
