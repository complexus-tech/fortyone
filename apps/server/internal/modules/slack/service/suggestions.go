package slack

import (
	"context"
	"strings"

	"github.com/google/uuid"
)

func (s *Service) handleBlockSuggestion(ctx context.Context, payload interactionPayload) (InteractionResponse, error) {
	if callbackID := strings.TrimSpace(payload.View.CallbackID); callbackID != "" && callbackID != "fortyone_create_task" {
		s.recordSuggestionDebug(ctx, payload, suggestionDebugInput{
			Outcome: "suggestion_skipped_invalid_callback",
			Reason:  "callback_id_not_supported",
		})
		return interactionOptionsResponse(nil)
	}

	source, err := parseSourceFromPrivateMetadata(payload.View.PrivateMetadata)
	if err != nil {
		s.recordSuggestionDebug(ctx, payload, suggestionDebugInput{
			Outcome: "suggestion_skipped_invalid_metadata",
			Reason:  err.Error(),
		})
		return interactionOptionsResponse(nil)
	}
	if strings.TrimSpace(source.SlackTeamID) == "" {
		source.SlackTeamID = strings.TrimSpace(payload.Team.ID)
	}
	source, err = interactionSourceForPayload(payload, source)
	if err != nil {
		s.recordSuggestionDebug(ctx, payload, suggestionDebugInput{
			Outcome:     "suggestion_skipped_actor_mismatch",
			Reason:      err.Error(),
			SlackTeamID: source.SlackTeamID,
		})
		return interactionOptionsResponse(nil)
	}

	slackWorkspace, err := s.repo.GetSlackWorkspaceByTeamID(ctx, source.SlackTeamID)
	if err != nil {
		s.recordSuggestionDebug(ctx, payload, suggestionDebugInput{
			Outcome:      "suggestion_skipped_workspace_not_found",
			Reason:       err.Error(),
			SlackTeamID:  source.SlackTeamID,
			WorkspaceID:  uuid.Nil,
			ResolvedTeam: uuid.Nil,
		})
		return interactionOptionsResponse(nil)
	}
	actorID, err := s.findLinkedInteractionActor(ctx, slackWorkspace.WorkspaceID, source)
	if err != nil {
		s.recordSuggestionDebug(ctx, payload, suggestionDebugInput{
			Outcome:     "suggestion_skipped_unlinked_actor",
			Reason:      err.Error(),
			SlackTeamID: source.SlackTeamID,
			WorkspaceID: slackWorkspace.WorkspaceID,
		})
		return interactionOptionsResponse(nil)
	}
	rawActionID := suggestionActionID(payload)
	query := suggestionQuery(payload)
	if rawActionID == modalActionTeamSelect {
		if blockID := suggestionBlockID(payload); blockID != "" && blockID != modalBlockTeam {
			s.recordSuggestionDebug(ctx, payload, suggestionDebugInput{
				Outcome:     "suggestion_skipped_unknown_action",
				Reason:      "team_action_block_id_not_valid",
				ActionID:    rawActionID,
				SlackTeamID: source.SlackTeamID,
				WorkspaceID: slackWorkspace.WorkspaceID,
			})
			return interactionOptionsResponse(nil)
		}
		teams, teamsErr := s.availableTeamsForSlackCreation(ctx, slackWorkspace.WorkspaceID, actorID)
		if teamsErr != nil {
			s.recordSuggestionDebug(ctx, payload, suggestionDebugInput{
				Outcome:     "suggestion_search_error_teams",
				Reason:      teamsErr.Error(),
				Query:       query,
				ActionID:    rawActionID,
				SlackTeamID: source.SlackTeamID,
				WorkspaceID: slackWorkspace.WorkspaceID,
			})
			return interactionOptionsResponse(nil)
		}
		options := slackTeamSuggestionOptions(teams, query)
		s.recordSuggestionDebug(ctx, payload, suggestionDebugInput{
			Outcome:     "suggestion_search_teams",
			Query:       query,
			ActionID:    rawActionID,
			SlackTeamID: source.SlackTeamID,
			WorkspaceID: slackWorkspace.WorkspaceID,
			ResultCount: len(options),
		})
		return interactionOptionsResponse(options)
	}

	_, teamID, actionCarriesTeam := modalDependentActionScope(rawActionID)
	if !actionCarriesTeam {
		teamID, err = s.resolveTeamIDForSuggestion(ctx, payload, slackWorkspace.WorkspaceID, actorID, source)
	}
	if err != nil || teamID == uuid.Nil {
		s.recordSuggestionDebug(ctx, payload, suggestionDebugInput{
			Outcome:      "suggestion_skipped_team_resolution_failed",
			Reason:       errorString(err, "selected team is missing"),
			SlackTeamID:  source.SlackTeamID,
			WorkspaceID:  slackWorkspace.WorkspaceID,
			ResolvedTeam: uuid.Nil,
		})
		return interactionOptionsResponse(nil)
	}
	if err := s.ensureTeamAvailableForSlackCreation(ctx, slackWorkspace.WorkspaceID, actorID, teamID); err != nil {
		s.recordSuggestionDebug(ctx, payload, suggestionDebugInput{
			Outcome:      "suggestion_skipped_team_not_available",
			Reason:       err.Error(),
			SlackTeamID:  source.SlackTeamID,
			WorkspaceID:  slackWorkspace.WorkspaceID,
			ResolvedTeam: teamID,
		})
		return interactionOptionsResponse(nil)
	}
	actionID, actionMatchesTeam := modalDependentActionBase(rawActionID, teamID)
	expectedBlockID := modalDependentBlockForAction(actionID)
	if !actionMatchesTeam || expectedBlockID == "" || !modalDependentBlockMatches(suggestionBlockID(payload), expectedBlockID, teamID) {
		s.recordSuggestionDebug(ctx, payload, suggestionDebugInput{
			Outcome:      "suggestion_skipped_unknown_action",
			Reason:       "action_or_block_id_not_valid_for_selected_team",
			ActionID:     rawActionID,
			SlackTeamID:  source.SlackTeamID,
			WorkspaceID:  slackWorkspace.WorkspaceID,
			ResolvedTeam: teamID,
		})
		return interactionOptionsResponse(nil)
	}

	if actionID != modalActionStatusSelect && len([]rune(query)) < 2 {
		s.recordSuggestionDebug(ctx, payload, suggestionDebugInput{
			Outcome:      "suggestion_skipped_query_too_short",
			Query:        query,
			ActionID:     rawActionID,
			SlackTeamID:  source.SlackTeamID,
			WorkspaceID:  slackWorkspace.WorkspaceID,
			ResolvedTeam: teamID,
		})
		return interactionOptionsResponse(nil)
	}

	const optionsLimit = 25
	switch actionID {
	case modalActionStatusSelect:
		statuses, statusesErr := s.repo.ListTeamStatuses(ctx, teamID)
		if statusesErr != nil {
			s.recordSuggestionDebug(ctx, payload, suggestionDebugInput{
				Outcome:      "suggestion_search_error_statuses",
				Reason:       statusesErr.Error(),
				Query:        query,
				ActionID:     modalActionStatusSelect,
				SlackTeamID:  source.SlackTeamID,
				WorkspaceID:  slackWorkspace.WorkspaceID,
				ResolvedTeam: teamID,
			})
			return interactionOptionsResponse(nil)
		}
		options := slackStatusSuggestionOptions(statuses, query)
		s.recordSuggestionDebug(ctx, payload, suggestionDebugInput{
			Outcome:      "suggestion_search_statuses",
			Query:        query,
			ActionID:     modalActionStatusSelect,
			SlackTeamID:  source.SlackTeamID,
			WorkspaceID:  slackWorkspace.WorkspaceID,
			ResolvedTeam: teamID,
			ResultCount:  len(options),
		})
		return interactionOptionsResponse(options)
	case modalActionAssigneeSelect:
		members, membersErr := s.repo.SearchTeamMembers(ctx, teamID, query, optionsLimit)
		if membersErr != nil {
			s.recordSuggestionDebug(ctx, payload, suggestionDebugInput{
				Outcome:      "suggestion_search_error_members",
				Reason:       membersErr.Error(),
				Query:        query,
				ActionID:     modalActionAssigneeSelect,
				SlackTeamID:  source.SlackTeamID,
				WorkspaceID:  slackWorkspace.WorkspaceID,
				ResolvedTeam: teamID,
			})
			return interactionOptionsResponse(nil)
		}
		options := make([]map[string]any, 0, len(members))
		for _, member := range members {
			options = append(options, toSlackOption(teamMemberDisplayName(member), member.UserID.String()))
		}
		s.recordSuggestionDebug(ctx, payload, suggestionDebugInput{
			Outcome:      "suggestion_search_members",
			Query:        query,
			ActionID:     modalActionAssigneeSelect,
			SlackTeamID:  source.SlackTeamID,
			WorkspaceID:  slackWorkspace.WorkspaceID,
			ResolvedTeam: teamID,
			ResultCount:  len(options),
		})
		return interactionOptionsResponse(options)
	case modalActionLabelsMultiSelect:
		labels, labelsErr := s.repo.SearchTeamLabels(ctx, slackWorkspace.WorkspaceID, teamID, query, optionsLimit)
		if labelsErr != nil {
			s.recordSuggestionDebug(ctx, payload, suggestionDebugInput{
				Outcome:      "suggestion_search_error_labels",
				Reason:       labelsErr.Error(),
				Query:        query,
				ActionID:     modalActionLabelsMultiSelect,
				SlackTeamID:  source.SlackTeamID,
				WorkspaceID:  slackWorkspace.WorkspaceID,
				ResolvedTeam: teamID,
			})
			return interactionOptionsResponse(nil)
		}
		options := make([]map[string]any, 0, len(labels))
		for _, label := range labels {
			options = append(options, toSlackOption(label.Name, label.ID.String()))
		}
		s.recordSuggestionDebug(ctx, payload, suggestionDebugInput{
			Outcome:      "suggestion_search_labels",
			Query:        query,
			ActionID:     modalActionLabelsMultiSelect,
			SlackTeamID:  source.SlackTeamID,
			WorkspaceID:  slackWorkspace.WorkspaceID,
			ResolvedTeam: teamID,
			ResultCount:  len(options),
		})
		return interactionOptionsResponse(options)
	case modalActionObjectiveSelect:
		objectives, objectivesErr := s.repo.SearchTeamObjectives(ctx, slackWorkspace.WorkspaceID, teamID, query, optionsLimit)
		if objectivesErr != nil {
			s.recordSuggestionDebug(ctx, payload, suggestionDebugInput{
				Outcome:      "suggestion_search_error_objectives",
				Reason:       objectivesErr.Error(),
				Query:        query,
				ActionID:     modalActionObjectiveSelect,
				SlackTeamID:  source.SlackTeamID,
				WorkspaceID:  slackWorkspace.WorkspaceID,
				ResolvedTeam: teamID,
			})
			return interactionOptionsResponse(nil)
		}
		options := make([]map[string]any, 0, len(objectives))
		for _, objective := range objectives {
			options = append(options, toSlackOption(objective.Name, objective.ID.String()))
		}
		s.recordSuggestionDebug(ctx, payload, suggestionDebugInput{
			Outcome:      "suggestion_search_objectives",
			Query:        query,
			ActionID:     modalActionObjectiveSelect,
			SlackTeamID:  source.SlackTeamID,
			WorkspaceID:  slackWorkspace.WorkspaceID,
			ResolvedTeam: teamID,
			ResultCount:  len(options),
		})
		return interactionOptionsResponse(options)
	default:
		s.recordSuggestionDebug(ctx, payload, suggestionDebugInput{
			Outcome:      "suggestion_skipped_unknown_action",
			Query:        query,
			ActionID:     rawActionID,
			SlackTeamID:  source.SlackTeamID,
			WorkspaceID:  slackWorkspace.WorkspaceID,
			ResolvedTeam: teamID,
		})
		return interactionOptionsResponse(nil)
	}
}

func (s *Service) resolveTeamIDForSuggestion(ctx context.Context, payload interactionPayload, workspaceID, actorID uuid.UUID, source requestSourceContext) (uuid.UUID, error) {
	if metadata, err := parseSlackModalPrivateMetadata(payload.View.PrivateMetadata); err == nil {
		if selectedFromMetadata := strings.TrimSpace(metadata.SelectedTeamID); selectedFromMetadata != "" {
			teamID, parseErr := uuid.Parse(selectedFromMetadata)
			if parseErr == nil && teamID != uuid.Nil {
				return teamID, nil
			}
		}
	}

	if selectedFromState := selectedTeamIDFromState(payload.View.State.Values); selectedFromState != "" {
		teamID, err := uuid.Parse(selectedFromState)
		if err == nil && teamID != uuid.Nil {
			return teamID, nil
		}
	}

	for _, block := range payload.View.Blocks {
		if strings.TrimSpace(block.BlockID) != modalBlockTeam {
			continue
		}
		value := strings.TrimSpace(block.Element.InitialOption.Value)
		if value == "" {
			continue
		}
		teamID, err := uuid.Parse(value)
		if err == nil && teamID != uuid.Nil {
			return teamID, nil
		}
	}

	teams, err := s.availableTeamsForSlackCreation(ctx, workspaceID, actorID)
	if err != nil {
		return uuid.Nil, err
	}
	if len(teams) == 0 {
		return uuid.Nil, ErrSlackNoTeamsAvailable
	}
	return teams[0].ID, nil
}
