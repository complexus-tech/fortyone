package chathttp

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	domain "github.com/complexus-tech/projects-api/internal/modules/chat/domain"
	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/stretchr/testify/require"
)

type avatarResolverStub struct {
	calls    []string
	expiries []time.Duration
	resolve  func(string) (string, error)
}

func (s *avatarResolverStub) ResolveProfileImageURL(_ context.Context, key string, expiry time.Duration) (string, error) {
	s.calls = append(s.calls, key)
	s.expiries = append(s.expiries, expiry)
	return s.resolve(key)
}

func TestChatResponsesResolvePrivateProfileImages(t *testing.T) {
	key := "profiles/person/avatar.webp"
	person := domain.Person{FullName: "A teammate", AvatarURL: &key}
	message := domain.Message{Author: person, ReplyAuthors: []domain.Person{person}}
	conversation := domain.Conversation{Members: []domain.Person{person}}
	values := map[string]any{
		"conversation":      conversation,
		"conversations":     domain.ConversationPage{Items: []domain.Conversation{conversation, conversation}},
		"members":           domain.MemberPage{Items: []domain.Person{person, person}},
		"message mutations": message,
		"history and replies": domain.MessagePage{
			Items: []domain.Message{message, message},
		},
		"search": domain.SearchPage{Items: []domain.Message{message, message}},
		"unread threads": domain.UnreadThreadPage{Items: []domain.UnreadThread{
			{Message: message, ConversationName: "Product", ConversationKind: "team"},
		}},
		"resources": domain.ResourcePage{
			Items: []domain.Resource{{Author: person}, {Author: person}},
		},
	}
	for name, value := range values {
		t.Run(name, func(t *testing.T) {
			before, err := json.Marshal(value)
			require.NoError(t, err)
			resolver := &avatarResolverStub{resolve: func(string) (string, error) {
				return "https://images.example.com/avatar.webp?signed=1", nil
			}}
			h := handlers{profileImages: resolver}
			w := httptest.NewRecorder()

			require.NoError(t, h.respond(t.Context(), w, value, nil, http.StatusOK))

			require.Equal(t, http.StatusOK, w.Code)
			require.NotContains(t, w.Body.String(), key)
			require.Contains(t, w.Body.String(), "https://images.example.com/avatar.webp?signed=1")
			require.Equal(t, []string{key}, resolver.calls, "repeated authors must only be signed once per response")
			require.Equal(t, []time.Duration{24 * time.Hour}, resolver.expiries)
			after, err := json.Marshal(value)
			require.NoError(t, err)
			require.JSONEq(t, string(before), string(after), "do not replace stored profile keys with expiring URLs")
		})
	}
}

func TestChatAvatarResolutionPreservesPublicImagesAndEmptyProfiles(t *testing.T) {
	publicURL, blank, raw := "https://images.example.com/public.webp", "  ", " profiles/avatar.webp "
	resolver := &avatarResolverStub{resolve: func(value string) (string, error) {
		if value == publicURL {
			return value, nil
		}
		return "https://images.example.com/private.webp?signed=1", nil
	}}
	h := handlers{profileImages: resolver}
	page := domain.MemberPage{Items: []domain.Person{
		{AvatarURL: &publicURL}, {AvatarURL: &blank}, {}, {AvatarURL: &raw},
	}}
	result := h.resolveResponseAvatars(t.Context(), page).(domain.MemberPage)
	require.Equal(t, publicURL, *result.Items[0].AvatarURL)
	require.Nil(t, result.Items[1].AvatarURL)
	require.Nil(t, result.Items[2].AvatarURL)
	require.Equal(t, "https://images.example.com/private.webp?signed=1", *result.Items[3].AvatarURL)
	require.Equal(t, []string{publicURL, "profiles/avatar.webp"}, resolver.calls)
}

func TestUnavailableChatAvatarsFallBackWithoutBreakingMessages(t *testing.T) {
	for _, scenario := range []string{"resolver missing", "signing failed", "empty result"} {
		t.Run(scenario, func(t *testing.T) {
			var logs bytes.Buffer
			h := handlers{log: logger.NewWithText(&logs, slog.LevelWarn, "test")}
			key := "profiles/private-user/avatar.webp"
			resolver := &avatarResolverStub{resolve: func(string) (string, error) {
				if scenario == "signing failed" {
					return "", errors.New("private signing details")
				}
				return "", nil
			}}
			if scenario != "resolver missing" {
				h.profileImages = resolver
			}
			person := domain.Person{FullName: "A teammate", AvatarURL: &key}
			message := domain.Message{Text: "Keep this message visible", Author: person, ReplyAuthors: []domain.Person{person}}
			w := httptest.NewRecorder()
			require.NoError(t, h.respond(t.Context(), w, message, nil, http.StatusOK))
			require.Equal(t, http.StatusOK, w.Code)
			require.Contains(t, w.Body.String(), "Keep this message visible")
			require.Contains(t, w.Body.String(), `"avatarUrl":null`)
			require.NotContains(t, w.Body.String(), key)
			require.NotContains(t, logs.String(), "private")
			if scenario != "resolver missing" {
				require.Equal(t, []string{key}, resolver.calls)
			}
			if scenario == "signing failed" {
				require.Contains(t, logs.String(), "failed to resolve chat profile image")
			}
		})
	}
}

func TestDeniedChatResponsesDoNotResolveProfileImages(t *testing.T) {
	key := "profiles/person/avatar.webp"
	resolver := &avatarResolverStub{resolve: func(string) (string, error) {
		t.Fatal("a denied response must not mint image access URLs")
		return "", nil
	}}
	h := handlers{profileImages: resolver}
	w := httptest.NewRecorder()
	require.NoError(t, h.respond(t.Context(), w, domain.Message{Author: domain.Person{AvatarURL: &key}}, domain.ErrForbidden, http.StatusOK))
	require.Equal(t, http.StatusForbidden, w.Code)
	require.Empty(t, resolver.calls)
}
