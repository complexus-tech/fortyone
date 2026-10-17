package messaging

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/uuid"
)

func TestDocumentResponseUsesOnlyBoundedDocumentTool(t *testing.T) {
	memberID := uuid.New()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var request responsesAPIRequest
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			t.Error(err)
		}
		if request.Store || len(request.Tools) != 1 || request.Tools[0].Name != "document_response" || request.ToolChoice == nil {
			t.Errorf("unsafe document tools: %+v", request.Tools)
		}
		if strings.Contains(request.Instructions, "list_team_members") || !strings.Contains(request.Instructions, "only for supplied eligible members") {
			t.Error("document instructions leaked story tools or omitted mention scope")
		}
		arguments, _ := json.Marshal(DocumentResponse{Reply: "I'll update the heading.", Operations: &DocumentOperations{Version: 1, Title: stringPointer("Revised title")}})
		_ = json.NewEncoder(w).Encode(map[string]any{"status": "completed", "output": []any{map[string]any{"type": "function_call", "call_id": "doc", "name": "document_response", "arguments": string(arguments)}}, "usage": map[string]int{"input_tokens": 10, "output_tokens": 5, "total_tokens": 15}})
	}))
	defer server.Close()
	assistant, err := NewOpenAIAssistant(OpenAIConfig{APIKey: "test", BaseURL: server.URL}, &assistantToolStub{})
	if err != nil {
		t.Fatal(err)
	}
	result, err := assistant.RespondDocument(context.Background(), DocumentRequest{UserID: uuid.New(), Title: "Original", Content: "Current text", Prompt: "Rename the document", CanEdit: true, Members: []DocumentMentionTarget{{ID: memberID, Name: "Person"}}})
	if err != nil {
		t.Fatal(err)
	}
	if result.Operations == nil || *result.Operations.Title != "Revised title" || result.Usage.TotalTokens != 15 {
		t.Fatalf("unexpected result %+v", result)
	}
}

func TestDocumentResponseRejectsUnauthorizedMalformedAndCrossSurfaceActions(t *testing.T) {
	for _, test := range []struct {
		name, arguments, tool string
		canEdit               bool
	}{
		{"read_only", `{"reply":"done","needsInput":false,"operations":{"version":1,"title":"Changed"}}`, "document_response", false},
		{"unknown_target", `{"reply":"done","needsInput":false,"operations":{"version":1,"title":"Changed","documentId":"foreign"}}`, "document_response", true},
		{"clarification_with_write", `{"reply":"which?","needsInput":true,"operations":{"version":1,"title":"Changed"}}`, "document_response", true},
		{"empty_replacement", `{"reply":"done","needsInput":false,"operations":{"version":1,"replacements":[{"before":"","after":"anything"}]}}`, "document_response", true},
		{"cross_surface", `{}`, "update_story", true},
	} {
		t.Run(test.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				_ = json.NewEncoder(w).Encode(map[string]any{"status": "completed", "output": []any{map[string]any{"type": "function_call", "call_id": "doc", "name": test.tool, "arguments": test.arguments}}})
			}))
			defer server.Close()
			assistant, err := NewOpenAIAssistant(OpenAIConfig{APIKey: "test", BaseURL: server.URL}, &assistantToolStub{})
			if err != nil {
				t.Fatal(err)
			}
			if _, err = assistant.RespondDocument(t.Context(), DocumentRequest{UserID: uuid.New(), Prompt: "Help", CanEdit: test.canEdit}); err == nil {
				t.Fatal("accepted unsafe document response")
			}
		})
	}
}

func stringPointer(value string) *string { return &value }
