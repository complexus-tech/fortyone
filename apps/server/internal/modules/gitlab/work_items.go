package gitlab

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/complexus-tech/projects-api/internal/platform/codehost"
)

type issueResponse struct {
	ID          int64  `json:"id"`
	IID         int64  `json:"iid"`
	ProjectID   int64  `json:"project_id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	State       string `json:"state"`
	WebURL      string `json:"web_url"`
}

type noteResponse struct {
	ID        int64     `json:"id"`
	Body      string    `json:"body"`
	CreatedAt time.Time `json:"created_at"`
	Author    struct {
		ID       int64  `json:"id"`
		Username string `json:"username"`
	} `json:"author"`
}

func (adapter *Adapter) CreateWorkItem(
	ctx context.Context,
	installation codehost.InstallationRef,
	command codehost.CreateWorkItem,
) (codehost.WorkItem, error) {
	if err := codehost.ValidateCreateWorkItem(command); err != nil {
		return codehost.WorkItem{}, err
	}
	projectID, err := numericRepositoryID(command.Repository)
	if err != nil {
		return codehost.WorkItem{}, err
	}
	var response issueResponse
	_, err = adapter.do(ctx, installation, http.MethodPost, fmt.Sprintf("projects/%d/issues", projectID), nil, map[string]string{
		"title":       command.Title,
		"description": command.Body,
	}, &response)
	if err != nil {
		return codehost.WorkItem{}, err
	}
	return mapWorkItem(command.Repository, response)
}

func (adapter *Adapter) AddComment(
	ctx context.Context,
	installation codehost.InstallationRef,
	command codehost.AddComment,
) (codehost.Comment, error) {
	if err := codehost.ValidateAddComment(command); err != nil {
		return codehost.Comment{}, err
	}
	projectID, err := numericRepositoryID(command.WorkItem.Repository)
	if err != nil {
		return codehost.Comment{}, err
	}
	var response noteResponse
	_, err = adapter.do(
		ctx,
		installation,
		http.MethodPost,
		fmt.Sprintf("projects/%d/issues/%d/notes", projectID, command.WorkItem.Number),
		nil,
		map[string]string{"body": command.Body},
		&response,
	)
	if err != nil {
		return codehost.Comment{}, err
	}
	return codehost.Comment{
		ExternalID:  strconv.FormatInt(response.ID, 10),
		WorkItem:    command.WorkItem,
		AuthorID:    strconv.FormatInt(response.Author.ID, 10),
		AuthorLogin: response.Author.Username,
		Body:        response.Body,
		WebURL:      fmt.Sprintf("%s/-/issues/%d#note_%d", strings.TrimSuffix(command.WorkItem.Repository.WebURL, "/"), command.WorkItem.Number, response.ID),
		CreatedAt:   response.CreatedAt,
	}, nil
}

func numericRepositoryID(repository codehost.RepositoryRef) (int64, error) {
	if err := codehost.ValidateRepository(repository); err != nil {
		return 0, err
	}
	projectID, err := strconv.ParseInt(repository.ExternalID, 10, 64)
	if err != nil || projectID <= 0 {
		return 0, codehost.ErrInvalidInput
	}
	return projectID, nil
}

func mapWorkItem(repository codehost.RepositoryRef, issue issueResponse) (codehost.WorkItem, error) {
	state, err := parseGitLabWorkItemState(issue.State)
	if err != nil {
		return codehost.WorkItem{}, err
	}
	return codehost.WorkItem{
		ExternalID: strconv.FormatInt(issue.ID, 10),
		Number:     issue.IID,
		Repository: repository,
		Title:      issue.Title,
		Body:       issue.Description,
		State:      state,
		WebURL:     issue.WebURL,
	}, nil
}

func parseGitLabWorkItemState(value string) (codehost.WorkItemState, error) {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "open", "opened", "reopened":
		return codehost.WorkItemStateOpen, nil
	case "close", "closed":
		return codehost.WorkItemStateClosed, nil
	default:
		return codehost.ParseWorkItemState(value)
	}
}

var (
	_ codehost.WorkItemWriter = (*Adapter)(nil)
	_ codehost.CommentWriter  = (*Adapter)(nil)
)
