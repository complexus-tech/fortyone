package workspacesgrp

import (
	"github.com/complexus-tech/projects-api/internal/core/stories"
	"github.com/google/uuid"
)

func seedStories(teamID uuid.UUID, userID uuid.UUID, statusID uuid.UUID) []stories.CoreNewStory {
	storyData := []struct {
		title       string
		description string
		priority    string
	}{
		{"Track your time", "Track your time", "Medium"},
		{"Create project documentation", "Create project documentation", "High"},
		{"Set up CI/CD pipeline", "Set up CI/CD pipeline", "High"},
		{"Implement user authentication", "Implement user authentication", "High"},
		{"Design user interface", "Design user interface", "Medium"},
	}

	s := make([]stories.CoreNewStory, len(storyData))
	for i, data := range storyData {
		desc := data.description
		descHTML := "<p>" + data.description + "</p>"
		s[i] = stories.CoreNewStory{
			Title:           data.title,
			Status:          &statusID,
			Description:     &desc,
			DescriptionHTML: &descHTML,
			Reporter:        &userID,
			Priority:        data.priority,
			Team:            teamID,
		}
	}

	return s
}
