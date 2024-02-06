package issuesgrp

import (
	"time"

	"github.com/complexus-tech/projects-api/internal/core/issues"
	"github.com/go-playground/validator/v10"
)

type AppIssue struct {
	ID          int        `json:"id" db:"id"`
	Title       string     `json:"title" db:"title"`
	Description string     `json:"description" db:"description"`
	CreatedAt   time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at" db:"updated_at"`
	DeletedAt   *time.Time `json:"deleted_at" db:"deleted_at"`
}

func toAppIssue(issue issues.Issue) AppIssue {
	return AppIssue{
		ID:          issue.ID,
		Title:       issue.Title,
		Description: issue.Description,
		CreatedAt:   issue.CreatedAt,
		UpdatedAt:   issue.UpdatedAt,
		DeletedAt:   issue.DeletedAt,
	}
}

func toAppIssues(issues []issues.Issue) []AppIssue {
	appIssues := make([]AppIssue, len(issues))
	for i, issue := range issues {
		appIssues[i] = toAppIssue(issue)
	}
	return appIssues
}

type AppNewIssue struct {
	Title       string `json:"title" validate:"required"`
	Description string `json:"description" validate:"required"`
}

func (a AppNewIssue) Validate() error {
	validate := validator.New(validator.WithRequiredStructEnabled())
	return validate.Struct(a)
}
