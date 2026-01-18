package epicsgrp

import (
	"github.com/complexus-tech/projects-api/internal/core/epics"
	"github.com/google/uuid"
)

// AppEpicList represents a epic in the application layer.
type AppEpicsList struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"title"`
	Description string    `json:"description"`
}

// toAppEpics converts a list of core epics to a list of application epics.
func toAppEpics(epics []epics.CoreEpic) []AppEpicsList {
	appEpics := make([]AppEpicsList, len(epics))
	for i, epic := range epics {
		appEpics[i] = AppEpicsList{
			ID:          epic.ID,
			Name:        epic.Name,
			Description: epic.Description,
		}
	}
	return appEpics
}
