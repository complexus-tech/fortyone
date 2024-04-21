package statesgrp

import (
	"github.com/complexus-tech/projects-api/internal/core/states"
	"github.com/google/uuid"
)

// AppStateList represents a state in the application layer.
type AppStatesList struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"title"`
	Description string    `json:"description"`
}

// toAppStates converts a list of core states to a list of application states.
func toAppStates(states []states.CoreState) []AppStatesList {
	appStates := make([]AppStatesList, len(states))
	for i, state := range states {
		appStates[i] = AppStatesList{
			ID:          state.ID,
			Name:        state.Name,
			Description: state.Description,
		}
	}
	return appStates
}
