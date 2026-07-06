package maya

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	stories "github.com/complexus-tech/projects-api/internal/modules/stories/service"
	"github.com/google/uuid"
)

func TestOpenAIAdvisorReturnsStructuredCandidateRecommendation(t *testing.T) {
	workspaceID := uuid.New()
	storyID := uuid.New()
	userID := uuid.New()
	startAt := time.Date(2026, 6, 15, 9, 0, 0, 0, time.UTC)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/responses" {
			t.Fatalf("expected responses path, got %q", r.URL.Path)
		}
		if r.Header.Get("Authorization") != "Bearer test-key" {
			t.Fatalf("expected bearer token, got %q", r.Header.Get("Authorization"))
		}
		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode request body: %v", err)
		}
		if body["model"] != "test-model" {
			t.Fatalf("expected model test-model, got %v", body["model"])
		}
		text, ok := body["text"].(map[string]any)
		if !ok {
			t.Fatalf("expected text config, got %#v", body["text"])
		}
		format, ok := text["format"].(map[string]any)
		if !ok || format["type"] != "json_schema" {
			t.Fatalf("expected json_schema text format, got %#v", text["format"])
		}
		requestPayload, err := json.Marshal(body["input"])
		if err != nil {
			t.Fatalf("marshal request input: %v", err)
		}
		if !strings.Contains(string(requestPayload), "Backend engineer") {
			t.Fatalf("expected role context in request payload, got %s", requestPayload)
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"id": "resp_test",
			"output": [{
				"content": [{
					"type": "output_text",
					"text": "{\"assigneeId\":\"` + userID.String() + `\",\"reason\":\"Best match for the integration area.\"}"
				}]
			}]
		}`))
	}))
	defer server.Close()

	advisor := NewOpenAIAdvisor(NewOpenAICompatibleClient(OpenAICompatibleConfig{
		APIKey:     "test-key",
		Model:      "test-model",
		BaseURL:    server.URL,
		HTTPClient: server.Client(),
	}))
	result, err := advisor.RecommendCandidate(context.Background(), CandidateRecommendationInput{
		WorkspaceID:     workspaceID,
		Story:           stories.CoreSingleStory{ID: storyID, Workspace: workspaceID, Title: "Fix GitHub sync"},
		DurationMinutes: 60,
		WindowStart:     startAt,
		WindowEnd:       startAt.Add(8 * time.Hour),
		Candidates: []CandidateRecommendation{
			{
				UserID:                userID,
				FullName:              "Ada Lovelace",
				TeamAIRoleTitle:       "Backend engineer",
				TeamAIRoleDescription: "Owns integrations, webhook reliability, and API sync work.",
				OpenStories:           2,
				EstimateTotal:         5,
				SlotStart:             startAt,
				SlotEnd:               startAt.Add(time.Hour),
			},
		},
	})

	if err != nil {
		t.Fatalf("RecommendCandidate returned error: %v", err)
	}
	if result.UserID != userID {
		t.Fatalf("expected user %s, got %s", userID, result.UserID)
	}
	if result.Reason != "Best match for the integration area." {
		t.Fatalf("unexpected reason %q", result.Reason)
	}
}

func TestOpenAIAdvisorReturnsStructuredBatchAssignmentRecommendations(t *testing.T) {
	workspaceID := uuid.New()
	firstStoryID := uuid.New()
	secondStoryID := uuid.New()
	firstUserID := uuid.New()
	secondUserID := uuid.New()
	startAt := time.Date(2026, 6, 15, 9, 0, 0, 0, time.UTC)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/responses" {
			t.Fatalf("expected responses path, got %q", r.URL.Path)
		}
		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode request body: %v", err)
		}
		requestPayload, err := json.Marshal(body["input"])
		if err != nil {
			t.Fatalf("marshal request input: %v", err)
		}
		if !strings.Contains(string(requestPayload), firstStoryID.String()) || !strings.Contains(string(requestPayload), secondStoryID.String()) {
			t.Fatalf("expected both stories in request payload, got %s", requestPayload)
		}
		if !strings.Contains(string(requestPayload), "Frontend engineer") || !strings.Contains(string(requestPayload), "Backend engineer") {
			t.Fatalf("expected candidate role context in request payload, got %s", requestPayload)
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"id": "resp_batch",
			"output": [{
				"content": [{
					"type": "output_text",
					"text": "{\"assignments\":[{\"storyId\":\"` + firstStoryID.String() + `\",\"assigneeId\":\"` + firstUserID.String() + `\",\"reason\":\"Best frontend fit.\"},{\"storyId\":\"` + secondStoryID.String() + `\",\"assigneeId\":\"` + secondUserID.String() + `\",\"reason\":\"Best backend fit.\"}]}"
				}]
			}]
		}`))
	}))
	defer server.Close()

	advisor := NewOpenAIAdvisor(NewOpenAICompatibleClient(OpenAICompatibleConfig{
		APIKey:     "test-key",
		Model:      "test-model",
		BaseURL:    server.URL,
		HTTPClient: server.Client(),
	}))
	results, err := advisor.RecommendAssignments(context.Background(), BatchAssignmentRecommendationInput{
		WorkspaceID: workspaceID,
		Stories: []BatchAssignmentStory{
			{ID: firstStoryID, Title: "Fix filter button spacing", Priority: "Medium"},
			{ID: secondStoryID, Title: "Add webhook retry endpoint", Priority: "High"},
		},
		Candidates: []CandidateRecommendation{
			{
				UserID:                firstUserID,
				FullName:              "Grace Hopper",
				TeamAIRoleTitle:       "Frontend engineer",
				TeamAIRoleDescription: "Owns UI, React, and design system work.",
				OpenStories:           3,
				EstimateTotal:         8,
				HasAvailableSlot:      true,
				SlotStart:             startAt,
				SlotEnd:               startAt.Add(time.Hour),
			},
			{
				UserID:                secondUserID,
				FullName:              "Katherine Johnson",
				TeamAIRoleTitle:       "Backend engineer",
				TeamAIRoleDescription: "Owns APIs, workers, and integrations.",
				OpenStories:           2,
				EstimateTotal:         5,
				HasAvailableSlot:      true,
				SlotStart:             startAt,
				SlotEnd:               startAt.Add(time.Hour),
			},
		},
	})

	if err != nil {
		t.Fatalf("RecommendAssignments returned error: %v", err)
	}
	if len(results.Assignments) != 2 {
		t.Fatalf("expected 2 assignments, got %d", len(results.Assignments))
	}
	if results.Assignments[0].StoryID != firstStoryID || results.Assignments[0].AssigneeID != firstUserID {
		t.Fatalf("unexpected first assignment %#v", results.Assignments[0])
	}
	if results.Assignments[1].StoryID != secondStoryID || results.Assignments[1].AssigneeID != secondUserID {
		t.Fatalf("unexpected second assignment %#v", results.Assignments[1])
	}
}
