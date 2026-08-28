package gitlab

import (
	"context"
	"encoding/json"
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/complexus-tech/projects-api/internal/platform/codehost"
)

type webhookPayload struct {
	ObjectKind string `json:"object_kind"`
	EventType  string `json:"event_type"`
	User       struct {
		ID       int64  `json:"id"`
		Username string `json:"username"`
	} `json:"user"`
	Project struct {
		ID                int64  `json:"id"`
		Name              string `json:"name"`
		PathWithNamespace string `json:"path_with_namespace"`
		WebURL            string `json:"web_url"`
		DefaultBranch     string `json:"default_branch"`
		VisibilityLevel   int    `json:"visibility_level"`
		Namespace         string `json:"namespace"`
	} `json:"project"`
	ObjectAttributes struct {
		ID           int64     `json:"id"`
		IID          int64     `json:"iid"`
		Title        string    `json:"title"`
		Description  string    `json:"description"`
		State        string    `json:"state"`
		URL          string    `json:"url"`
		Action       string    `json:"action"`
		Note         string    `json:"note"`
		NoteableType string    `json:"noteable_type"`
		CreatedAt    time.Time `json:"created_at"`
	} `json:"object_attributes"`
	Issue struct {
		ID          int64  `json:"id"`
		IID         int64  `json:"iid"`
		Title       string `json:"title"`
		Description string `json:"description"`
		State       string `json:"state"`
		URL         string `json:"url"`
	} `json:"issue"`
}

func (adapter *Adapter) NormalizeWebhook(
	_ context.Context,
	deliveryID, eventType string,
	body []byte,
) (codehost.NormalizedEvent, error) {
	deliveryID = strings.TrimSpace(deliveryID)
	eventType = strings.TrimSpace(eventType)
	if deliveryID == "" || len(body) == 0 {
		return codehost.NormalizedEvent{}, codehost.ErrInvalidInput
	}
	if eventType != "Issue Hook" && eventType != "Note Hook" {
		return codehost.NormalizedEvent{}, codehost.ErrCapabilityUnsupported
	}
	var payload webhookPayload
	if err := json.Unmarshal(body, &payload); err != nil {
		return codehost.NormalizedEvent{}, errors.Join(codehost.ErrInvalidInput, err)
	}
	repository, err := mapWebhookRepository(payload)
	if err != nil {
		return codehost.NormalizedEvent{}, err
	}
	event := codehost.NormalizedEvent{
		Provider:             ProviderKey,
		DeliveryID:           deliveryID,
		Action:               strings.TrimSpace(payload.ObjectAttributes.Action),
		ExternalRepositoryID: strconv.FormatInt(payload.Project.ID, 10),
		ExternalActorID:      strconv.FormatInt(payload.User.ID, 10),
	}
	switch eventType {
	case "Issue Hook":
		if payload.ObjectKind != "issue" || payload.ObjectAttributes.IID <= 0 {
			return codehost.NormalizedEvent{}, codehost.ErrInvalidInput
		}
		item, err := webhookWorkItem(repository, payload.ObjectAttributes.ID, payload.ObjectAttributes.IID,
			payload.ObjectAttributes.Title, payload.ObjectAttributes.Description,
			payload.ObjectAttributes.State, payload.ObjectAttributes.URL)
		if err != nil {
			return codehost.NormalizedEvent{}, err
		}
		event.Kind = codehost.EventWorkItemChanged
		event.WorkItem = &item
	case "Note Hook":
		if payload.ObjectKind != "note" || payload.ObjectAttributes.NoteableType != "Issue" ||
			payload.ObjectAttributes.ID <= 0 || payload.Issue.IID <= 0 {
			return codehost.NormalizedEvent{}, codehost.ErrCapabilityUnsupported
		}
		item, err := webhookWorkItem(repository, payload.Issue.ID, payload.Issue.IID,
			payload.Issue.Title, payload.Issue.Description, payload.Issue.State, payload.Issue.URL)
		if err != nil {
			return codehost.NormalizedEvent{}, err
		}
		event.Kind = codehost.EventCommentCreated
		if event.Action == "" {
			event.Action = "create"
		}
		event.Comment = &codehost.Comment{
			ExternalID:  strconv.FormatInt(payload.ObjectAttributes.ID, 10),
			WorkItem:    item,
			AuthorID:    strconv.FormatInt(payload.User.ID, 10),
			AuthorLogin: payload.User.Username,
			Body:        payload.ObjectAttributes.Note,
			WebURL:      payload.ObjectAttributes.URL,
			CreatedAt:   payload.ObjectAttributes.CreatedAt,
		}
	}
	return event, nil
}

func mapWebhookRepository(payload webhookPayload) (codehost.RepositoryRef, error) {
	project := projectResponse{
		ID:                payload.Project.ID,
		Name:              payload.Project.Name,
		PathWithNamespace: payload.Project.PathWithNamespace,
		WebURL:            payload.Project.WebURL,
		DefaultBranch:     payload.Project.DefaultBranch,
		Visibility:        "private",
	}
	project.Namespace.FullPath = payload.Project.Namespace
	if payload.Project.VisibilityLevel >= 20 {
		project.Visibility = "public"
	}
	repository := mapRepository(project)
	if err := codehost.ValidateRepository(repository); err != nil {
		return codehost.RepositoryRef{}, err
	}
	return repository, nil
}

func webhookWorkItem(
	repository codehost.RepositoryRef,
	id, iid int64,
	title, body, stateValue, webURL string,
) (codehost.WorkItem, error) {
	if id <= 0 || iid <= 0 {
		return codehost.WorkItem{}, codehost.ErrInvalidInput
	}
	state, err := parseGitLabWorkItemState(stateValue)
	if err != nil {
		return codehost.WorkItem{}, err
	}
	return codehost.WorkItem{
		ExternalID: strconv.FormatInt(id, 10),
		Number:     iid,
		Repository: repository,
		Title:      title,
		Body:       body,
		State:      state,
		WebURL:     webURL,
	}, nil
}

var _ codehost.WebhookNormalizer = (*Adapter)(nil)
