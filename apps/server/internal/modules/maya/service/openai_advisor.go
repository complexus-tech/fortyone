package maya

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
)

const (
	defaultAIBaseURL = "https://api.openai.com/v1"
	defaultAIModel   = "gpt-5-nano-2025-08-07"
)

type OpenAICompatibleConfig struct {
	APIKey     string
	Model      string
	BaseURL    string
	HTTPClient *http.Client
}

type OpenAICompatibleClient struct {
	apiKey     string
	model      string
	baseURL    string
	httpClient *http.Client
}

func NewOpenAICompatibleClient(cfg OpenAICompatibleConfig) *OpenAICompatibleClient {
	model := strings.TrimSpace(cfg.Model)
	if model == "" {
		model = defaultAIModel
	}
	baseURL := strings.TrimRight(strings.TrimSpace(cfg.BaseURL), "/")
	if baseURL == "" {
		baseURL = defaultAIBaseURL
	}
	client := cfg.HTTPClient
	if client == nil {
		client = &http.Client{Timeout: 20 * time.Second}
	}
	return &OpenAICompatibleClient{
		apiKey:     strings.TrimSpace(cfg.APIKey),
		model:      model,
		baseURL:    baseURL,
		httpClient: client,
	}
}

func (c *OpenAICompatibleClient) Enabled() bool {
	return c != nil && c.apiKey != ""
}

func (c *OpenAICompatibleClient) postJSON(ctx context.Context, path string, body any) ([]byte, error) {
	if !c.Enabled() {
		return nil, errors.New("ai client is not configured")
	}
	payload, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("marshal ai request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+path, bytes.NewReader(payload))
	if err != nil {
		return nil, fmt.Errorf("create ai request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Content-Type", "application/json")

	res, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("call ai provider: %w", err)
	}
	defer res.Body.Close()

	data, err := io.ReadAll(io.LimitReader(res.Body, 1<<20))
	if err != nil {
		return nil, fmt.Errorf("read ai response: %w", err)
	}
	if res.StatusCode < http.StatusOK || res.StatusCode >= http.StatusMultipleChoices {
		return nil, fmt.Errorf("ai provider returned %s: %s", res.Status, strings.TrimSpace(string(data)))
	}
	return data, nil
}

type OpenAIAdvisor struct {
	client *OpenAICompatibleClient
}

func NewOpenAIAdvisor(client *OpenAICompatibleClient) *OpenAIAdvisor {
	return &OpenAIAdvisor{client: client}
}

func (a *OpenAIAdvisor) RecommendCandidate(ctx context.Context, input CandidateRecommendationInput) (CandidateRecommendationResult, error) {
	if a == nil || !a.client.Enabled() {
		return CandidateRecommendationResult{}, errors.New("openai advisor is not configured")
	}

	data, err := a.client.postJSON(ctx, "/responses", responsesRequest{
		Model: a.client.model,
		Input: []responsesMessage{
			{
				Role: "developer",
				Content: []responsesContent{
					{
						Type: "input_text",
						Text: mayaAssignmentSystemPrompt(),
					},
				},
			},
			{
				Role: "user",
				Content: []responsesContent{
					{
						Type: "input_text",
						Text: mustMarshalString(toOpenAIAdvisorPrompt(input)),
					},
				},
			},
		},
		Text: responsesText{
			Format: mayaAssignmentResponseFormat(),
		},
	})
	if err != nil {
		return CandidateRecommendationResult{}, err
	}

	content, err := responseOutputText(data)
	if err != nil {
		return CandidateRecommendationResult{}, err
	}
	var output openAIAdvisorOutput
	if err := json.Unmarshal([]byte(content), &output); err != nil {
		return CandidateRecommendationResult{}, fmt.Errorf("decode ai advisor output: %w", err)
	}
	userID, err := uuid.Parse(strings.TrimSpace(output.AssigneeID))
	if err != nil {
		return CandidateRecommendationResult{}, fmt.Errorf("parse ai advisor assignee id: %w", err)
	}
	return CandidateRecommendationResult{
		UserID: userID,
		Reason: strings.TrimSpace(output.Reason),
	}, nil
}

func (a *OpenAIAdvisor) RecommendAssignments(ctx context.Context, input BatchAssignmentRecommendationInput) (BatchAssignmentRecommendationResult, error) {
	if a == nil || !a.client.Enabled() {
		return BatchAssignmentRecommendationResult{}, errors.New("openai advisor is not configured")
	}

	data, err := a.client.postJSON(ctx, "/responses", responsesRequest{
		Model: a.client.model,
		Input: []responsesMessage{
			{
				Role: "developer",
				Content: []responsesContent{
					{
						Type: "input_text",
						Text: mayaBatchAssignmentSystemPrompt(),
					},
				},
			},
			{
				Role: "user",
				Content: []responsesContent{
					{
						Type: "input_text",
						Text: mustMarshalString(toOpenAIBatchAssignmentPrompt(input)),
					},
				},
			},
		},
		Text: responsesText{
			Format: mayaBatchAssignmentResponseFormat(),
		},
	})
	if err != nil {
		return BatchAssignmentRecommendationResult{}, err
	}

	content, err := responseOutputText(data)
	if err != nil {
		return BatchAssignmentRecommendationResult{}, err
	}
	var output openAIBatchAdvisorOutput
	if err := json.Unmarshal([]byte(content), &output); err != nil {
		return BatchAssignmentRecommendationResult{}, fmt.Errorf("decode ai batch advisor output: %w", err)
	}
	assignments := make([]BatchAssignmentRecommendation, 0, len(output.Assignments))
	for _, item := range output.Assignments {
		storyID, err := uuid.Parse(strings.TrimSpace(item.StoryID))
		if err != nil {
			return BatchAssignmentRecommendationResult{}, fmt.Errorf("parse ai batch advisor story id: %w", err)
		}
		assigneeID, err := uuid.Parse(strings.TrimSpace(item.AssigneeID))
		if err != nil {
			return BatchAssignmentRecommendationResult{}, fmt.Errorf("parse ai batch advisor assignee id: %w", err)
		}
		assignments = append(assignments, BatchAssignmentRecommendation{
			StoryID:    storyID,
			AssigneeID: assigneeID,
			Reason:     strings.TrimSpace(item.Reason),
		})
	}
	return BatchAssignmentRecommendationResult{Assignments: assignments}, nil
}

func mayaAssignmentSystemPrompt() string {
	return strings.Join([]string{
		"You are Maya, a project management scheduling assistant.",
		"Choose the best assignee from the provided candidates only.",
		"Prioritize ownership fit, team role context, recent work fit inferred from the story, workload, urgency, and available calendar time.",
		"Use candidate role titles and role descriptions as strong signals when they are present.",
		"Never invent users, dates, or ids.",
		"Return only valid JSON matching the provided schema.",
	}, " ")
}

func mayaBatchAssignmentSystemPrompt() string {
	return strings.Join([]string{
		"You are Maya, a project management scheduling assistant.",
		"Assign each provided story to exactly one provided candidate.",
		"Prioritize ownership fit, team role context, recent work fit inferred from each story, workload, urgency, and available calendar time.",
		"Use candidate role titles and role descriptions as strong signals when they are present.",
		"Never invent users, stories, dates, or ids.",
		"Return only valid JSON matching the provided schema.",
	}, " ")
}

func mayaAssignmentResponseFormat() map[string]any {
	return map[string]any{
		"type":   "json_schema",
		"name":   "maya_assignment_recommendation",
		"strict": true,
		"schema": map[string]any{
			"type":                 "object",
			"additionalProperties": false,
			"properties": map[string]any{
				"assigneeId": map[string]any{
					"type":        "string",
					"description": "The selected candidate user id. Must match one of the supplied candidate user ids.",
				},
				"reason": map[string]any{
					"type":        "string",
					"description": "A concise product-facing explanation for why this assignee is the best fit.",
				},
			},
			"required": []string{"assigneeId", "reason"},
		},
	}
}

func mayaBatchAssignmentResponseFormat() map[string]any {
	return map[string]any{
		"type":   "json_schema",
		"name":   "maya_batch_assignment_recommendations",
		"strict": true,
		"schema": map[string]any{
			"type":                 "object",
			"additionalProperties": false,
			"properties": map[string]any{
				"assignments": map[string]any{
					"type": "array",
					"items": map[string]any{
						"type":                 "object",
						"additionalProperties": false,
						"properties": map[string]any{
							"storyId": map[string]any{
								"type":        "string",
								"description": "The story id being assigned. Must match one of the supplied story ids.",
							},
							"assigneeId": map[string]any{
								"type":        "string",
								"description": "The selected candidate user id. Must match one of the supplied candidate user ids.",
							},
							"reason": map[string]any{
								"type":        "string",
								"description": "A concise product-facing explanation for why this assignee is the best fit.",
							},
						},
						"required": []string{"storyId", "assigneeId", "reason"},
					},
				},
			},
			"required": []string{"assignments"},
		},
	}
}

func responseOutputText(data []byte) (string, error) {
	var response responsesResponse
	if err := json.Unmarshal(data, &response); err != nil {
		return "", fmt.Errorf("decode ai response: %w", err)
	}
	for _, output := range response.Output {
		for _, content := range output.Content {
			if strings.TrimSpace(content.Text) != "" {
				return content.Text, nil
			}
		}
	}
	return "", errors.New("ai provider returned no output text")
}

func toOpenAIBatchAssignmentPrompt(input BatchAssignmentRecommendationInput) map[string]any {
	storiesPayload := make([]map[string]any, 0, len(input.Stories))
	for _, story := range input.Stories {
		storiesPayload = append(storiesPayload, map[string]any{
			"id":              story.ID,
			"title":           story.Title,
			"description":     story.Description,
			"priority":        story.Priority,
			"estimateValue":   story.EstimateValue,
			"estimateLabel":   story.EstimateLabel,
			"durationMinutes": story.DurationMinutes,
		})
	}

	candidates := make([]map[string]any, 0, len(input.Candidates))
	for _, candidate := range input.Candidates {
		candidatePayload := map[string]any{
			"userId":           candidate.UserID,
			"fullName":         candidate.FullName,
			"username":         candidate.Username,
			"roleTitle":        candidate.TeamAIRoleTitle,
			"roleDescription":  candidate.TeamAIRoleDescription,
			"openItems":        candidate.OpenStories,
			"estimateTotal":    candidate.EstimateTotal,
			"hasAvailableSlot": candidate.HasAvailableSlot,
			"availableSlot":    nil,
		}
		if candidate.HasAvailableSlot {
			candidatePayload["availableSlot"] = map[string]any{
				"start": candidate.SlotStart.Format(time.RFC3339),
				"end":   candidate.SlotEnd.Format(time.RFC3339),
			}
		}
		candidates = append(candidates, candidatePayload)
	}

	return map[string]any{
		"workspaceId": input.WorkspaceID,
		"stories":     storiesPayload,
		"candidates":  candidates,
	}
}

func toOpenAIAdvisorPrompt(input CandidateRecommendationInput) map[string]any {
	description := ""
	if input.Story.Description != nil {
		description = *input.Story.Description
	}
	candidates := make([]map[string]any, 0, len(input.Candidates))
	for _, candidate := range input.Candidates {
		candidatePayload := map[string]any{
			"userId":           candidate.UserID,
			"fullName":         candidate.FullName,
			"username":         candidate.Username,
			"roleTitle":        candidate.TeamAIRoleTitle,
			"roleDescription":  candidate.TeamAIRoleDescription,
			"openItems":        candidate.OpenStories,
			"estimateTotal":    candidate.EstimateTotal,
			"hasAvailableSlot": candidate.HasAvailableSlot,
			"availableSlot":    nil,
		}
		if candidate.HasAvailableSlot {
			candidatePayload["availableSlot"] = map[string]any{
				"start": candidate.SlotStart.Format(time.RFC3339),
				"end":   candidate.SlotEnd.Format(time.RFC3339),
			}
		}
		candidates = append(candidates, candidatePayload)
	}
	return map[string]any{
		"story": map[string]any{
			"id":              input.Story.ID,
			"title":           input.Story.Title,
			"description":     description,
			"priority":        input.Story.Priority,
			"estimateValue":   input.Story.EstimateValue,
			"estimateLabel":   input.Story.EstimateLabel,
			"durationMinutes": input.DurationMinutes,
		},
		"planningWindow": map[string]any{
			"start": input.WindowStart.Format(time.RFC3339),
			"end":   input.WindowEnd.Format(time.RFC3339),
		},
		"candidates": candidates,
	}
}

func mustMarshalString(value any) string {
	data, err := json.Marshal(value)
	if err != nil {
		return "{}"
	}
	return string(data)
}

type responsesRequest struct {
	Model string             `json:"model"`
	Input []responsesMessage `json:"input"`
	Text  responsesText      `json:"text"`
}

type responsesMessage struct {
	Role    string             `json:"role"`
	Content []responsesContent `json:"content"`
}

type responsesContent struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

type responsesText struct {
	Format map[string]any `json:"format"`
}

type responsesResponse struct {
	Output []struct {
		Content []struct {
			Type string `json:"type"`
			Text string `json:"text"`
		} `json:"content"`
	} `json:"output"`
}

type openAIAdvisorOutput struct {
	AssigneeID string `json:"assigneeId"`
	Reason     string `json:"reason"`
}

type openAIBatchAdvisorOutput struct {
	Assignments []struct {
		StoryID    string `json:"storyId"`
		AssigneeID string `json:"assigneeId"`
		Reason     string `json:"reason"`
	} `json:"assignments"`
}
