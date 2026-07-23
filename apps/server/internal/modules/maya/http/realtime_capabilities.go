package mayahttp

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	activities "github.com/complexus-tech/projects-api/internal/modules/activities/service"
	objectives "github.com/complexus-tech/projects-api/internal/modules/objectives/service"
	reports "github.com/complexus-tech/projects-api/internal/modules/reports/service"
	sprints "github.com/complexus-tech/projects-api/internal/modules/sprints/service"
	states "github.com/complexus-tech/projects-api/internal/modules/states/service"
	stories "github.com/complexus-tech/projects-api/internal/modules/stories/service"
	teams "github.com/complexus-tech/projects-api/internal/modules/teams/service"
	users "github.com/complexus-tech/projects-api/internal/modules/users/service"
	"github.com/google/uuid"
)

type realtimeFeedbackMatch struct {
	item   AppRealtimeVoiceFeedbackItem
	id     uuid.UUID
	teamID uuid.UUID
}

func (h *Handlers) executeNavigate(ctx context.Context, workspaceID, userID uuid.UUID, rawArgs json.RawMessage) (AppRealtimeToolResponse, error) {
	var args AppRealtimeNavigateArguments
	if err := decodeRealtimeArguments(rawArgs, &args, "navigate"); err != nil {
		return AppRealtimeToolResponse{}, err
	}
	args.TargetType = strings.TrimSpace(args.TargetType)
	args.Reference = strings.TrimSpace(args.Reference)
	args.TeamName = strings.TrimSpace(args.TeamName)
	args.Route = strings.TrimSpace(args.Route)

	staticPaths := map[string]string{
		"my-work":       "/my-work",
		"summary":       "/summary",
		"analytics":     "/analytics",
		"sprints":       "/sprints",
		"notifications": "/notifications",
		"settings":      "/settings",
		"roadmaps":      "/roadmaps",
		"billing":       "/settings/workspace/billing",
	}
	if path, ok := staticPaths[args.TargetType]; ok {
		return realtimeNavigationResponse(path, args.TargetType), nil
	}

	teamList, err := h.teams.List(ctx, workspaceID, userID)
	if err != nil {
		return AppRealtimeToolResponse{}, fmt.Errorf("list teams for navigation: %w", err)
	}
	switch args.TargetType {
	case "team":
		team := resolveRealtimeTeam(teamList, firstNonEmpty(args.TeamName, args.Reference))
		if team == nil {
			return realtimeTeamClarification(teamList, "Ask which team to open."), nil
		}
		route := args.Route
		allowedRoutes := map[string]bool{
			"": true, "stories": true, "sprints": true, "objectives": true,
			"backlog": true, "feedback": true, "requests": true,
		}
		if !allowedRoutes[route] {
			return AppRealtimeToolResponse{Success: false, Error: "That team page is not available for voice navigation."}, nil
		}
		if route == "" {
			route = "stories"
		}
		return realtimeNavigationResponse(fmt.Sprintf("/teams/%s/%s", team.ID, route), "team"), nil
	case "story":
		story, response, err := h.resolveRealtimeStory(ctx, workspaceID, userID, args.Reference)
		if err != nil || response != nil {
			return responseOrEmpty(response), err
		}
		return realtimeNavigationResponse(fmt.Sprintf("/story/%s", story.ID), "story"), nil
	case "feedback":
		if strings.TrimSpace(args.TeamName) != "" && resolveRealtimeTeam(teamList, args.TeamName) == nil {
			return realtimeTeamClarification(teamList, "Ask which team's feedback to open."), nil
		}
		matches, err := h.findRealtimeFeedback(ctx, workspaceID, userID, teamList, args.TeamName, args.Reference, "all", 10)
		if err != nil {
			return AppRealtimeToolResponse{}, err
		}
		exact := exactRealtimeFeedbackMatches(matches, args.Reference)
		if len(exact) == 1 {
			return realtimeNavigationResponse(fmt.Sprintf("/teams/%s/feedback/%s", exact[0].teamID, exact[0].id), "feedback"), nil
		}
		return AppRealtimeToolResponse{
			Success: false, FeedbackItems: feedbackItems(matches),
			Message: "Ask which feedback item the user meant.",
		}, nil
	case "objective":
		objective, team, response, err := h.resolveRealtimeObjective(ctx, workspaceID, userID, args.Reference, args.TeamName)
		if err != nil || response != nil {
			return responseOrEmpty(response), err
		}
		return realtimeNavigationResponse(fmt.Sprintf("/teams/%s/objectives/%s", team.ID, objective.ID), "objective"), nil
	case "sprint":
		sprint, team, response, err := h.resolveRealtimeSprint(ctx, workspaceID, userID, args.Reference, args.TeamName)
		if err != nil || response != nil {
			return responseOrEmpty(response), err
		}
		return realtimeNavigationResponse(fmt.Sprintf("/teams/%s/sprints/%s/stories", team.ID, sprint.ID), "sprint"), nil
	case "user-profile":
		members, err := h.users.List(ctx, workspaceID, users.CoreListUsersFilter{Search: args.Reference, Limit: 10})
		if err != nil {
			return AppRealtimeToolResponse{}, fmt.Errorf("find user for navigation: %w", err)
		}
		matches := resolveRealtimeMemberMatches(members, args.Reference)
		if len(matches) != 1 {
			return AppRealtimeToolResponse{
				Success: false,
				Members: toRealtimeVoiceMembers(members),
				Message: "Ask which person to open.",
			}, nil
		}
		return realtimeNavigationResponse(fmt.Sprintf("/profile/%s", matches[0].ID), "user-profile"), nil
	default:
		return AppRealtimeToolResponse{Success: false, Error: "That destination is not available."}, nil
	}
}

func executeSetTheme(rawArgs json.RawMessage) AppRealtimeToolResponse {
	var args AppRealtimeSetThemeArguments
	if err := decodeRealtimeArguments(rawArgs, &args, "set_theme"); err != nil {
		return AppRealtimeToolResponse{Success: false, Error: err.Error()}
	}
	theme := strings.ToLower(strings.TrimSpace(args.Theme))
	if theme != "light" && theme != "dark" && theme != "system" && theme != "toggle" {
		return AppRealtimeToolResponse{Success: false, Error: "Theme must be light, dark, system, or toggle."}
	}
	return AppRealtimeToolResponse{
		Success:      true,
		Message:      "Theme updated.",
		ClientAction: &AppRealtimeClientAction{Type: "theme", Theme: theme},
	}
}

func (h *Handlers) executeGetStory(ctx context.Context, workspaceID, userID uuid.UUID, rawArgs json.RawMessage) (AppRealtimeToolResponse, error) {
	var args AppRealtimeStoryArguments
	if err := decodeRealtimeArguments(rawArgs, &args, "get_story"); err != nil {
		return AppRealtimeToolResponse{}, err
	}
	story, response, err := h.resolveRealtimeStory(ctx, workspaceID, userID, args.Reference)
	if err != nil || response != nil {
		return responseOrEmpty(response), err
	}
	voiceStory, err := h.toRealtimeVoiceSingleStory(ctx, workspaceID, userID, story)
	if err != nil {
		return AppRealtimeToolResponse{}, err
	}
	return AppRealtimeToolResponse{
		Success: true,
		Story:   &voiceStory,
		Message: fmt.Sprintf("Found %s: %s.", voiceStory.Reference, voiceStory.Title),
	}, nil
}

func (h *Handlers) executeUpdateStory(ctx context.Context, workspaceID, userID, sessionID uuid.UUID, rawArgs json.RawMessage) (AppRealtimeToolResponse, error) {
	var args AppRealtimeUpdateStoryArguments
	if err := decodeRealtimeArguments(rawArgs, &args, "update_story"); err != nil {
		return AppRealtimeToolResponse{}, err
	}
	args.Reference = strings.TrimSpace(args.Reference)
	args.Title = strings.TrimSpace(args.Title)
	args.Status = strings.TrimSpace(args.Status)
	args.Priority = strings.TrimSpace(args.Priority)
	args.AssigneeName = strings.TrimSpace(args.AssigneeName)
	args.StartDate = strings.TrimSpace(args.StartDate)
	args.EndDate = strings.TrimSpace(args.EndDate)
	args.SprintName = strings.TrimSpace(args.SprintName)
	args.ObjectiveName = strings.TrimSpace(args.ObjectiveName)

	story, response, err := h.resolveRealtimeStory(ctx, workspaceID, userID, args.Reference)
	if err != nil || response != nil {
		return responseOrEmpty(response), err
	}
	updates, summary, response, err := h.resolveRealtimeStoryUpdates(ctx, workspaceID, userID, story, args)
	if err != nil || response != nil {
		return responseOrEmpty(response), err
	}
	if len(updates) == 0 {
		return AppRealtimeToolResponse{Success: false, Message: "Ask what should change on the story."}, nil
	}

	isConfirmed := args.Confirmed
	providedToken := args.ConfirmationToken
	args.Confirmed = false
	args.ConfirmationToken = ""
	confirmationInput := struct {
		StoryID uuid.UUID
		Updates map[string]any
	}{StoryID: story.ID, Updates: updates}
	expectedToken, err := h.confirmationToken(sessionID, "update_story", confirmationInput)
	if err != nil {
		return AppRealtimeToolResponse{}, err
	}
	if !isConfirmed {
		return AppRealtimeToolResponse{
			Success:              false,
			RequiresConfirmation: true,
			ConfirmationToken:    expectedToken,
			Message:              fmt.Sprintf("Ask the user to confirm updating %s: %s.", storyReference(story.TeamCode, story.SequenceID), strings.Join(summary, ", ")),
		}, nil
	}
	valid, err := h.validateConfirmationToken(sessionID, "update_story", confirmationInput, providedToken)
	if err != nil {
		return AppRealtimeToolResponse{}, err
	}
	if !valid {
		return changedConfirmationResponse(expectedToken), nil
	}
	if err := h.stories.Update(ctx, story.ID, workspaceID, updates); err != nil {
		return AppRealtimeToolResponse{}, fmt.Errorf("update story: %w", err)
	}
	h.invalidateStoryListCaches(ctx, workspaceID)
	updated, err := h.stories.Get(ctx, story.ID, workspaceID)
	if err != nil {
		return AppRealtimeToolResponse{}, fmt.Errorf("read updated story: %w", err)
	}
	voiceStory, err := h.toRealtimeVoiceSingleStory(ctx, workspaceID, userID, updated)
	if err != nil {
		return AppRealtimeToolResponse{}, err
	}
	return AppRealtimeToolResponse{
		Success: true,
		Story:   &voiceStory,
		Message: fmt.Sprintf("Updated %s.", voiceStory.Reference),
	}, nil
}

func (h *Handlers) executeStoryComments(ctx context.Context, workspaceID, userID, sessionID uuid.UUID, rawArgs json.RawMessage) (AppRealtimeToolResponse, error) {
	args := AppRealtimeCommentsArguments{Action: "list", Limit: 10}
	if err := decodeRealtimeArguments(rawArgs, &args, "story_comments"); err != nil {
		return AppRealtimeToolResponse{}, err
	}
	story, response, err := h.resolveRealtimeStory(ctx, workspaceID, userID, args.Reference)
	if err != nil || response != nil {
		return responseOrEmpty(response), err
	}
	reference := storyReference(story.TeamCode, story.SequenceID)
	switch strings.ToLower(strings.TrimSpace(args.Action)) {
	case "list":
		commentList, _, err := h.stories.GetComments(ctx, story.ID, 1, clampLimit(args.Limit, 10))
		if err != nil {
			return AppRealtimeToolResponse{}, fmt.Errorf("list story comments: %w", err)
		}
		userIDs := make([]uuid.UUID, 0, len(commentList))
		for _, comment := range commentList {
			userIDs = append(userIDs, comment.UserID)
		}
		authors, err := h.users.GetUsersByIDs(ctx, userIDs)
		if err != nil {
			return AppRealtimeToolResponse{}, fmt.Errorf("resolve comment authors: %w", err)
		}
		authorsByID := make(map[uuid.UUID]users.CoreUser, len(authors))
		for _, author := range authors {
			authorsByID[author.ID] = author
		}
		result := make([]AppRealtimeVoiceComment, 0, len(commentList))
		for _, comment := range commentList {
			result = append(result, AppRealtimeVoiceComment{
				Author:    displayUserName(authorsByID[comment.UserID]),
				Comment:   strings.TrimSpace(comment.Comment),
				CreatedAt: comment.CreatedAt,
			})
		}
		return AppRealtimeToolResponse{
			Success: true, Comments: result, Count: len(result),
			Message: fmt.Sprintf("Found %d comment%s on %s.", len(result), pluralSuffix(len(result)), reference),
		}, nil
	case "add":
		args.Comment = strings.TrimSpace(args.Comment)
		if args.Comment == "" {
			return AppRealtimeToolResponse{Success: false, Message: "Ask what comment to add."}, nil
		}
		isConfirmed := args.Confirmed
		providedToken := args.ConfirmationToken
		confirmationInput := struct {
			StoryID uuid.UUID
			Comment string
		}{StoryID: story.ID, Comment: args.Comment}
		expectedToken, err := h.confirmationToken(sessionID, "story_comments", confirmationInput)
		if err != nil {
			return AppRealtimeToolResponse{}, err
		}
		if !isConfirmed {
			return AppRealtimeToolResponse{
				Success: false, RequiresConfirmation: true, ConfirmationToken: expectedToken,
				Message: fmt.Sprintf("Ask the user to confirm adding this comment to %s: %q.", reference, args.Comment),
			}, nil
		}
		valid, err := h.validateConfirmationToken(sessionID, "story_comments", confirmationInput, providedToken)
		if err != nil {
			return AppRealtimeToolResponse{}, err
		}
		if !valid {
			return changedConfirmationResponse(expectedToken), nil
		}
		comment, err := h.stories.CreateComment(ctx, workspaceID, stories.CoreNewComment{
			StoryID: story.ID, UserID: userID, Comment: args.Comment,
		})
		if err != nil {
			return AppRealtimeToolResponse{}, fmt.Errorf("add story comment: %w", err)
		}
		return AppRealtimeToolResponse{
			Success:  true,
			Comments: []AppRealtimeVoiceComment{{Author: "You", Comment: comment.Comment, CreatedAt: comment.CreatedAt}},
			Message:  fmt.Sprintf("Added the comment to %s.", reference),
		}, nil
	default:
		return AppRealtimeToolResponse{Success: false, Error: "Comment action must be list or add."}, nil
	}
}

func (h *Handlers) executeSprints(ctx context.Context, workspaceID, userID uuid.UUID, rawArgs json.RawMessage) (AppRealtimeToolResponse, error) {
	args := AppRealtimeSprintArguments{Action: "list_running", Limit: 10}
	if err := decodeRealtimeArguments(rawArgs, &args, "sprints"); err != nil {
		return AppRealtimeToolResponse{}, err
	}
	action := strings.ToLower(strings.TrimSpace(args.Action))
	if action == "list_running" {
		sprintList, err := h.sprints.Running(ctx, workspaceID, userID)
		if err != nil {
			return AppRealtimeToolResponse{}, fmt.Errorf("list running sprints: %w", err)
		}
		teamList, err := h.teams.List(ctx, workspaceID, userID)
		if err != nil {
			return AppRealtimeToolResponse{}, fmt.Errorf("list sprint teams: %w", err)
		}
		result := toRealtimeVoiceSprints(sprintList, indexTeamsByID(teamList), clampLimit(args.Limit, 10))
		return AppRealtimeToolResponse{
			Success: true, Sprints: result, Count: len(result),
			Message: fmt.Sprintf("Found %d running sprint%s.", len(result), pluralSuffix(len(result))),
		}, nil
	}
	if action != "get_summary" {
		return AppRealtimeToolResponse{Success: false, Error: "Sprint action must be list_running or get_summary."}, nil
	}
	sprint, team, response, err := h.resolveRealtimeSprint(ctx, workspaceID, userID, args.Name, args.TeamName)
	if err != nil || response != nil {
		return responseOrEmpty(response), err
	}
	analytics, err := h.sprints.GetAnalytics(ctx, sprint.ID, workspaceID)
	if err != nil {
		return AppRealtimeToolResponse{}, fmt.Errorf("get sprint summary: %w", err)
	}
	voiceSprint := toRealtimeVoiceSprint(sprint, team.Name)
	voiceSprint.Status = analytics.Overview.Status
	voiceSprint.CompletionPercentage = analytics.Overview.CompletionPercentage
	return AppRealtimeToolResponse{
		Success: true, Sprints: []AppRealtimeVoiceSprint{voiceSprint}, Count: 1,
		Message: fmt.Sprintf("%s is %d%% complete and %s.", sprint.Name, voiceSprint.CompletionPercentage, strings.ReplaceAll(voiceSprint.Status, "_", " ")),
	}, nil
}

func (h *Handlers) executeWorkload(ctx context.Context, workspaceID, userID uuid.UUID, rawArgs json.RawMessage) (AppRealtimeToolResponse, error) {
	args := AppRealtimeWorkloadArguments{Limit: 5}
	if err := decodeRealtimeArguments(rawArgs, &args, "workload"); err != nil {
		return AppRealtimeToolResponse{}, err
	}
	filters := reports.ReportFilters{}
	if strings.TrimSpace(args.TeamName) != "" {
		teamList, err := h.teams.List(ctx, workspaceID, userID)
		if err != nil {
			return AppRealtimeToolResponse{}, fmt.Errorf("list workload teams: %w", err)
		}
		team := resolveRealtimeTeam(teamList, args.TeamName)
		if team == nil {
			return realtimeTeamClarification(teamList, "Ask which team's workload to summarize."), nil
		}
		filters.TeamIDs = []uuid.UUID{team.ID}
	}
	analysis, err := h.reports.GetWorkloadAnalysis(ctx, workspaceID, filters)
	if err != nil {
		return AppRealtimeToolResponse{}, fmt.Errorf("get workload analysis: %w", err)
	}
	members := analysis.Risks.OverloadedMembers
	if len(members) == 0 {
		members = analysis.Risks.OverdueMembers
	}
	limit := clampLimit(args.Limit, 5)
	if len(members) > limit {
		members = members[:limit]
	}
	atRisk := make([]AppRealtimeVoiceWorkloadMember, 0, len(members))
	for _, member := range members {
		atRisk = append(atRisk, AppRealtimeVoiceWorkloadMember{
			Name:        displayWorkloadMemberName(member.FullName, member.Username),
			OpenStories: member.OpenStories, OverdueStories: member.OverdueStories,
			UrgentStories: member.UrgentStories, EstimateTotal: member.EstimateTotal,
		})
	}
	workload := AppRealtimeVoiceWorkload{
		TotalOpenStories:   analysis.Summary.TotalOpenStories,
		TotalEstimate:      analysis.Summary.TotalEstimate,
		OverdueStories:     analysis.Summary.OverdueStories,
		UnassignedStories:  analysis.Summary.UnassignedStories,
		UnestimatedStories: analysis.Summary.UnestimatedStories,
		AtRiskMembers:      atRisk,
	}
	return AppRealtimeToolResponse{
		Success: true, Workload: &workload,
		Message: fmt.Sprintf("There are %d open stories, %d overdue, and %d unassigned.", workload.TotalOpenStories, workload.OverdueStories, workload.UnassignedStories),
	}, nil
}

func (h *Handlers) executeRecentActivity(ctx context.Context, workspaceID, userID uuid.UUID, rawArgs json.RawMessage) (AppRealtimeToolResponse, error) {
	args := AppRealtimeActivityArguments{Days: 7, Limit: 10}
	if err := decodeRealtimeArguments(rawArgs, &args, "recent_activity"); err != nil {
		return AppRealtimeToolResponse{}, err
	}
	if args.Days <= 0 || args.Days > 90 {
		args.Days = 7
	}
	limit := clampLimit(args.Limit, 10)
	now := h.now().UTC()
	activityList, err := h.activities.GetActivities(ctx, userID, limit, workspaceID, activities.ActivityFilters{
		StartDate: now.AddDate(0, 0, -args.Days),
		EndDate:   now,
	})
	if err != nil {
		return AppRealtimeToolResponse{}, fmt.Errorf("get recent activity: %w", err)
	}
	result := make([]AppRealtimeVoiceActivity, 0, len(activityList))
	storyReferences := make(map[uuid.UUID]string, len(activityList))
	for _, activity := range activityList {
		actor := firstNonEmpty(strings.TrimSpace(activity.User.FullName), strings.TrimSpace(activity.User.Username), "Someone")
		reference, ok := storyReferences[activity.StoryID]
		if !ok && activity.StoryID != uuid.Nil {
			if story, err := h.stories.Get(ctx, activity.StoryID, workspaceID); err == nil {
				reference = storyReference(story.TeamCode, story.SequenceID)
			}
			storyReferences[activity.StoryID] = reference
		}
		result = append(result, AppRealtimeVoiceActivity{
			Actor: actor, StoryRef: reference, Type: activity.Type, Field: activity.Field,
			Value: activity.CurrentValue, CreatedAt: activity.CreatedAt,
		})
	}
	return AppRealtimeToolResponse{
		Success: true, Activities: result, Count: len(result),
		Message: fmt.Sprintf("Found %d recent activit%s.", len(result), activityPluralEnding(len(result))),
	}, nil
}

func (h *Handlers) executeNotifications(ctx context.Context, workspaceID, userID, sessionID uuid.UUID, rawArgs json.RawMessage) (AppRealtimeToolResponse, error) {
	args := AppRealtimeNotificationsArguments{Action: "list", Limit: 10}
	if err := decodeRealtimeArguments(rawArgs, &args, "notifications"); err != nil {
		return AppRealtimeToolResponse{}, err
	}
	action := strings.ToLower(strings.TrimSpace(args.Action))
	switch action {
	case "list":
		items, err := h.notifications.List(ctx, userID, workspaceID, "", clampLimit(args.Limit, 10), 0)
		if err != nil {
			return AppRealtimeToolResponse{}, fmt.Errorf("list notifications: %w", err)
		}
		result := make([]AppRealtimeVoiceNotification, 0, len(items))
		for _, item := range items {
			result = append(result, AppRealtimeVoiceNotification{
				Title: item.Title, Type: item.Type, EntityType: item.EntityType,
				IsRead: item.ReadAt != nil, CreatedAt: item.CreatedAt,
			})
		}
		unread, err := h.notifications.GetUnreadCount(ctx, userID, workspaceID)
		if err != nil {
			return AppRealtimeToolResponse{}, fmt.Errorf("get unread notifications: %w", err)
		}
		return AppRealtimeToolResponse{
			Success: true, Notifications: result, Count: len(result),
			Message: fmt.Sprintf("You have %d unread notification%s.", unread, pluralSuffix(unread)),
		}, nil
	case "unread_count":
		unread, err := h.notifications.GetUnreadCount(ctx, userID, workspaceID)
		if err != nil {
			return AppRealtimeToolResponse{}, fmt.Errorf("get unread notifications: %w", err)
		}
		return AppRealtimeToolResponse{Success: true, Count: unread, Message: fmt.Sprintf("You have %d unread notification%s.", unread, pluralSuffix(unread))}, nil
	case "mark_read", "mark_all_read":
		confirmationInput := struct {
			Action string
			Title  string
		}{Action: action, Title: strings.TrimSpace(args.Title)}
		expectedToken, err := h.confirmationToken(sessionID, "notifications", confirmationInput)
		if err != nil {
			return AppRealtimeToolResponse{}, err
		}
		if !args.Confirmed {
			message := "Ask the user to confirm marking all notifications as read."
			if action == "mark_read" {
				message = fmt.Sprintf("Ask the user to confirm marking the notification %q as read.", confirmationInput.Title)
			}
			return AppRealtimeToolResponse{
				Success: false, RequiresConfirmation: true,
				ConfirmationToken: expectedToken, Message: message,
			}, nil
		}
		valid, err := h.validateConfirmationToken(sessionID, "notifications", confirmationInput, args.ConfirmationToken)
		if err != nil {
			return AppRealtimeToolResponse{}, err
		}
		if !valid {
			return changedConfirmationResponse(expectedToken), nil
		}
		if action == "mark_all_read" {
			if err := h.notifications.MarkAllAsRead(ctx, userID, workspaceID); err != nil {
				return AppRealtimeToolResponse{}, fmt.Errorf("mark all notifications read: %w", err)
			}
			return AppRealtimeToolResponse{Success: true, Message: "Marked all notifications as read."}, nil
		}
		items, err := h.notifications.List(ctx, userID, workspaceID, confirmationInput.Title, 20, 0)
		if err != nil {
			return AppRealtimeToolResponse{}, fmt.Errorf("find notification: %w", err)
		}
		matches := make([]uuid.UUID, 0, len(items))
		for _, item := range items {
			if normalizeName(item.Title) == normalizeName(confirmationInput.Title) {
				matches = append(matches, item.ID)
			}
		}
		if len(matches) != 1 {
			return AppRealtimeToolResponse{Success: false, Message: "I could not uniquely identify that notification. Ask for its exact title."}, nil
		}
		if err := h.notifications.MarkAsRead(ctx, matches[0], userID); err != nil {
			return AppRealtimeToolResponse{}, fmt.Errorf("mark notification read: %w", err)
		}
		return AppRealtimeToolResponse{Success: true, Message: "Marked the notification as read."}, nil
	default:
		return AppRealtimeToolResponse{Success: false, Error: "Notification action must be list, unread_count, mark_read, or mark_all_read."}, nil
	}
}

func (h *Handlers) executeCustomerFeedback(ctx context.Context, workspaceID, userID uuid.UUID, rawArgs json.RawMessage) (AppRealtimeToolResponse, error) {
	args := AppRealtimeFeedbackArguments{Action: "list", Status: "active", Limit: 10}
	if err := decodeRealtimeArguments(rawArgs, &args, "customer_feedback"); err != nil {
		return AppRealtimeToolResponse{}, err
	}
	action := strings.ToLower(strings.TrimSpace(args.Action))
	if action != "list" && action != "get" {
		return AppRealtimeToolResponse{Success: false, Error: "Feedback action must be list or get."}, nil
	}
	teamList, err := h.teams.List(ctx, workspaceID, userID)
	if err != nil {
		return AppRealtimeToolResponse{}, fmt.Errorf("list feedback teams: %w", err)
	}
	if strings.TrimSpace(args.TeamName) != "" && resolveRealtimeTeam(teamList, args.TeamName) == nil {
		return realtimeTeamClarification(teamList, "Ask which team's feedback to check."), nil
	}
	limit := clampLimit(args.Limit, 10)
	matches, err := h.findRealtimeFeedback(ctx, workspaceID, userID, teamList, args.TeamName, args.Title, firstNonEmpty(strings.TrimSpace(args.Status), "active"), limit)
	if err != nil {
		return AppRealtimeToolResponse{}, err
	}
	if action == "get" {
		exact := exactRealtimeFeedbackMatches(matches, args.Title)
		if len(exact) == 1 {
			match := exact[0]
			return AppRealtimeToolResponse{
				Success: true, FeedbackItems: []AppRealtimeVoiceFeedbackItem{match.item}, Count: 1,
				Message: fmt.Sprintf("Found feedback %q.", match.item.Title),
			}, nil
		}
		if len(matches) != 1 {
			return AppRealtimeToolResponse{Success: false, FeedbackItems: feedbackItems(matches), Message: "Ask which feedback item the user meant."}, nil
		}
	}
	items := feedbackItems(matches)
	return AppRealtimeToolResponse{
		Success: true, FeedbackItems: items, Count: len(items),
		Message: fmt.Sprintf("Found %d feedback item%s.", len(items), pluralSuffix(len(items))),
	}, nil
}

func (h *Handlers) findRealtimeFeedback(ctx context.Context, workspaceID, userID uuid.UUID, teamList []teams.CoreTeam, teamName, title, status string, limit int) ([]realtimeFeedbackMatch, error) {
	selectedTeams := teamList
	if strings.TrimSpace(teamName) != "" {
		team := resolveRealtimeTeam(teamList, teamName)
		if team == nil {
			return nil, nil
		}
		selectedTeams = []teams.CoreTeam{*team}
	}
	matches := make([]realtimeFeedbackMatch, 0, limit)
	for _, team := range selectedTeams {
		page, err := h.feedback.ListTeamItems(ctx, workspaceID, team.ID, userID, status, strings.TrimSpace(title), 1, limit)
		if err != nil {
			return nil, fmt.Errorf("list customer feedback: %w", err)
		}
		for _, item := range page.Items {
			matches = append(matches, realtimeFeedbackMatch{
				id: item.ID, teamID: team.ID,
				item: AppRealtimeVoiceFeedbackItem{
					Title: item.Title, Status: item.Status, Team: team.Name,
					Description: item.Description, Author: item.AuthorName, VoteCount: item.VoteCount,
					CommentCount: item.CommentCount, CreatedAt: item.CreatedAt,
				},
			})
			if item.RoadmapSummary != nil {
				matches[len(matches)-1].item.RoadmapSummary = *item.RoadmapSummary
			}
			if len(matches) >= limit {
				break
			}
		}
		if len(matches) >= limit {
			break
		}
	}
	return matches, nil
}

func exactRealtimeFeedbackMatches(matches []realtimeFeedbackMatch, title string) []realtimeFeedbackMatch {
	exact := make([]realtimeFeedbackMatch, 0, len(matches))
	for _, match := range matches {
		if normalizeName(match.item.Title) == normalizeName(title) {
			exact = append(exact, match)
		}
	}
	return exact
}

func (h *Handlers) executeWorkspaceBriefing(ctx context.Context, workspaceID, userID uuid.UUID, rawArgs json.RawMessage) (AppRealtimeToolResponse, error) {
	args := AppRealtimeWorkspaceBriefingArguments{Days: 30}
	if err := decodeRealtimeArguments(rawArgs, &args, "workspace_briefing"); err != nil {
		return AppRealtimeToolResponse{}, err
	}
	if args.Days <= 0 || args.Days > 365 {
		args.Days = 30
	}
	end := h.now().UTC()
	start := end.AddDate(0, 0, -args.Days)
	filters := reports.ReportFilters{StartDate: &start, EndDate: &end}
	overview, err := h.reports.GetWorkspaceOverview(ctx, workspaceID, filters)
	if err != nil {
		return AppRealtimeToolResponse{}, fmt.Errorf("get workspace overview: %w", err)
	}
	pulse, err := h.reports.GetPulseReport(ctx, workspaceID, filters)
	if err != nil {
		return AppRealtimeToolResponse{}, fmt.Errorf("get workspace pulse: %w", err)
	}
	feedbackSummaries, err := h.feedback.ListTeamSummaries(ctx, workspaceID, userID)
	if err != nil {
		return AppRealtimeToolResponse{}, fmt.Errorf("get feedback summary: %w", err)
	}
	feedbackItems := 0
	unreadFeedback := 0
	for _, summary := range feedbackSummaries {
		feedbackItems += summary.TotalCount
		unreadFeedback += summary.UnreadCount
	}
	briefing := AppRealtimeVoiceBriefing{
		TotalStories:      overview.Metrics.TotalStories,
		CompletedStories:  overview.Metrics.CompletedStories,
		ActiveObjectives:  overview.Metrics.ActiveObjectives,
		ActiveSprints:     overview.Metrics.ActiveSprints,
		TeamMembers:       overview.Metrics.TotalTeamMembers,
		OverdueStories:    pulse.Summary.OverdueStories,
		BlockedStories:    pulse.Summary.BlockedStories,
		AtRiskSprints:     pulse.Summary.AtRiskSprints,
		AtRiskObjectives:  pulse.Summary.AtRiskObjectives,
		FeedbackItems:     feedbackItems,
		UnreadFeedback:    unreadFeedback,
		OverloadedMembers: pulse.Summary.OverloadedMembers,
	}
	return AppRealtimeToolResponse{
		Success: true, Briefing: &briefing,
		Message: fmt.Sprintf("In the last %d days: %d stories completed, %d overdue, %d blocked, %d at-risk objectives, and %d unread feedback items.", args.Days, briefing.CompletedStories, briefing.OverdueStories, briefing.BlockedStories, briefing.AtRiskObjectives, briefing.UnreadFeedback),
	}, nil
}

func (h *Handlers) resolveRealtimeStory(ctx context.Context, workspaceID, userID uuid.UUID, reference string) (stories.CoreSingleStory, *AppRealtimeToolResponse, error) {
	teamList, err := h.teams.List(ctx, workspaceID, userID)
	if err != nil {
		return stories.CoreSingleStory{}, nil, fmt.Errorf("list teams while resolving story: %w", err)
	}
	id, _, response, err := h.resolveRealtimeStoryLink(ctx, workspaceID, userID, teamList, nil, reference, "work item")
	if err != nil || response != nil {
		return stories.CoreSingleStory{}, response, err
	}
	if id == nil {
		return stories.CoreSingleStory{}, &AppRealtimeToolResponse{Success: false, NeedsStoryReference: true, Message: "Ask which work item the user meant."}, nil
	}
	story, err := h.stories.Get(ctx, *id, workspaceID)
	if err != nil {
		return stories.CoreSingleStory{}, nil, fmt.Errorf("get story: %w", err)
	}
	return story, nil, nil
}

func (h *Handlers) toRealtimeVoiceSingleStory(ctx context.Context, workspaceID, userID uuid.UUID, story stories.CoreSingleStory) (AppRealtimeVoiceStory, error) {
	team, err := h.teams.GetByID(ctx, story.Team, workspaceID, userID)
	if err != nil {
		return AppRealtimeVoiceStory{}, fmt.Errorf("get story team: %w", err)
	}
	statuses, err := h.states.TeamList(ctx, workspaceID, story.Team)
	if err != nil {
		return AppRealtimeVoiceStory{}, fmt.Errorf("get story statuses: %w", err)
	}
	var statusName *AppRealtimeVoiceStatus
	if story.Status != nil {
		for _, status := range statuses {
			if status.ID == *story.Status {
				statusName = toRealtimeVoiceStatus(status)
				break
			}
		}
	}
	assignee := ""
	if story.Assignee != nil {
		if user, err := h.users.GetUser(ctx, *story.Assignee); err == nil {
			assignee = displayUserName(user)
		}
	}
	description := ""
	if story.Description != nil {
		description = strings.TrimSpace(*story.Description)
	}
	sprintName := ""
	if story.SprintSummary != nil {
		sprintName = story.SprintSummary.Name
	}
	objectiveName := ""
	if story.Objective != nil {
		if objective, err := h.objectives.Get(ctx, *story.Objective, workspaceID); err == nil {
			objectiveName = objective.Name
		}
	}
	return AppRealtimeVoiceStory{
		Reference: storyReference(team.Code, story.SequenceID), Title: story.Title, Description: description,
		Priority: story.Priority, EstimateLabel: story.EstimateLabel,
		EstimateValue: story.EstimateValue, Team: team.Name, Assignee: assignee,
		Sprint: sprintName, Objective: objectiveName, Status: statusName,
		StartDate: story.StartDate, EndDate: story.EndDate,
		CompletedAt: story.CompletedAt,
	}, nil
}

func (h *Handlers) resolveRealtimeStoryUpdates(ctx context.Context, workspaceID, userID uuid.UUID, story stories.CoreSingleStory, args AppRealtimeUpdateStoryArguments) (map[string]any, []string, *AppRealtimeToolResponse, error) {
	updates := make(map[string]any)
	summary := make([]string, 0, 7)
	if args.Title != "" {
		updates["title"] = args.Title
		summary = append(summary, fmt.Sprintf("title to %q", args.Title))
	}
	if args.Priority != "" {
		priority := normalizePriority(args.Priority)
		if _, ok := realtimeStoryPriorities[priority]; !ok {
			return nil, nil, &AppRealtimeToolResponse{Success: false, Message: "Ask for a priority of No Priority, Low, Medium, High, or Urgent."}, nil
		}
		updates["priority"] = priority
		summary = append(summary, "priority to "+priority)
	}
	if args.Status != "" {
		statuses, err := h.states.TeamList(ctx, workspaceID, story.Team)
		if err != nil {
			return nil, nil, nil, fmt.Errorf("list story statuses: %w", err)
		}
		matched := resolveRealtimeStatus(statuses, args.Status)
		if matched == nil {
			return nil, nil, &AppRealtimeToolResponse{Success: false, Message: "Ask which exact team status to use."}, nil
		}
		updates["status_id"] = matched.ID
		summary = append(summary, "status to "+matched.Name)
	}
	if args.Unassign && (args.AssignToMe || args.AssigneeName != "") {
		return nil, nil, &AppRealtimeToolResponse{Success: false, Message: "Ask whether to unassign the story or assign it to someone."}, nil
	}
	if args.Unassign {
		updates["assignee_id"] = nil
		summary = append(summary, "remove the assignee")
	} else if args.AssignToMe || args.AssigneeName != "" {
		team, err := h.teams.GetByID(ctx, story.Team, workspaceID, userID)
		if err != nil {
			return nil, nil, nil, fmt.Errorf("get story team: %w", err)
		}
		assigneeID, assigneeName, response, err := h.resolveRealtimeAssignee(ctx, workspaceID, userID, &team, AppRealtimeCreateTaskArguments{
			AssigneeName: args.AssigneeName, AssignToMe: args.AssignToMe,
		})
		if err != nil || response != nil {
			return nil, nil, response, err
		}
		updates["assignee_id"] = assigneeID
		summary = append(summary, "assignee to "+assigneeName)
	}
	if args.ClearEstimate && args.EstimateValue != nil {
		return nil, nil, &AppRealtimeToolResponse{Success: false, Message: "Ask whether to clear the estimate or set a new one."}, nil
	}
	if args.ClearEstimate {
		updates["estimate_unit"] = nil
		summary = append(summary, "clear the estimate")
	} else if args.EstimateValue != nil {
		updates["estimate_unit"] = args.EstimateValue
		summary = append(summary, fmt.Sprintf("estimate to %d", *args.EstimateValue))
	}
	currentUser, err := h.users.GetUser(ctx, userID)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("get current user for dates: %w", err)
	}
	loc := userLocation(currentUser)
	now := h.now().In(loc)
	if args.ClearStartDate && args.StartDate != "" {
		return nil, nil, &AppRealtimeToolResponse{Success: false, Message: "Ask whether to clear the start date or set a new one."}, nil
	}
	if args.ClearStartDate {
		updates["start_date"] = nil
		summary = append(summary, "clear the start date")
	} else if args.StartDate != "" {
		value, err := parseRealtimeDate(args.StartDate, loc, now)
		if err != nil {
			return nil, nil, &AppRealtimeToolResponse{Success: false, Message: "Ask the user to clarify the start date."}, nil
		}
		updates["start_date"] = value
		summary = append(summary, "start date to "+args.StartDate)
	}
	if args.ClearEndDate && args.EndDate != "" {
		return nil, nil, &AppRealtimeToolResponse{Success: false, Message: "Ask whether to clear the due date or set a new one."}, nil
	}
	if args.ClearEndDate {
		updates["end_date"] = nil
		summary = append(summary, "clear the due date")
	} else if args.EndDate != "" {
		value, err := parseRealtimeDate(args.EndDate, loc, now)
		if err != nil {
			return nil, nil, &AppRealtimeToolResponse{Success: false, Message: "Ask the user to clarify the due date."}, nil
		}
		updates["end_date"] = value
		summary = append(summary, "due date to "+args.EndDate)
	}
	if args.ClearSprint && args.SprintName != "" {
		return nil, nil, &AppRealtimeToolResponse{Success: false, Message: "Ask whether to remove the sprint or set a new one."}, nil
	}
	if args.ClearSprint {
		updates["sprint_id"] = nil
		summary = append(summary, "remove the sprint")
	} else if args.SprintName != "" {
		team, err := h.teams.GetByID(ctx, story.Team, workspaceID, userID)
		if err != nil {
			return nil, nil, nil, fmt.Errorf("get story team for sprint: %w", err)
		}
		sprint, _, response, err := h.resolveRealtimeSprint(ctx, workspaceID, userID, args.SprintName, team.Name)
		if err != nil || response != nil {
			return nil, nil, response, err
		}
		updates["sprint_id"] = sprint.ID
		summary = append(summary, "sprint to "+sprint.Name)
	}
	if args.ClearObjective && args.ObjectiveName != "" {
		return nil, nil, &AppRealtimeToolResponse{Success: false, Message: "Ask whether to remove the objective or set a new one."}, nil
	}
	if args.ClearObjective {
		updates["objective_id"] = nil
		summary = append(summary, "remove the objective")
	} else if args.ObjectiveName != "" {
		team, err := h.teams.GetByID(ctx, story.Team, workspaceID, userID)
		if err != nil {
			return nil, nil, nil, fmt.Errorf("get story team for objective: %w", err)
		}
		objective, _, response, err := h.resolveRealtimeObjective(ctx, workspaceID, userID, args.ObjectiveName, team.Name)
		if err != nil || response != nil {
			return nil, nil, response, err
		}
		updates["objective_id"] = objective.ID
		summary = append(summary, "objective to "+objective.Name)
	}
	return updates, summary, nil, nil
}

func (h *Handlers) resolveRealtimeSprint(ctx context.Context, workspaceID, userID uuid.UUID, name, teamName string) (sprints.CoreSprint, teams.CoreTeam, *AppRealtimeToolResponse, error) {
	teamList, err := h.teams.List(ctx, workspaceID, userID)
	if err != nil {
		return sprints.CoreSprint{}, teams.CoreTeam{}, nil, fmt.Errorf("list sprint teams: %w", err)
	}
	var teamID *uuid.UUID
	if strings.TrimSpace(teamName) != "" {
		team := resolveRealtimeTeam(teamList, teamName)
		if team == nil {
			return sprints.CoreSprint{}, teams.CoreTeam{}, ptr(realtimeTeamClarification(teamList, "Ask which team's sprint the user meant.")), nil
		}
		teamID = &team.ID
	}
	filters := map[string]any{}
	if teamID != nil {
		filters["team_id"] = *teamID
	}
	sprintList, err := h.sprints.List(ctx, workspaceID, userID, filters)
	if err != nil {
		return sprints.CoreSprint{}, teams.CoreTeam{}, nil, fmt.Errorf("list sprints: %w", err)
	}
	normalized := normalizeName(name)
	matches := make([]int, 0, len(sprintList))
	for i, sprint := range sprintList {
		if normalizeName(sprint.Name) == normalized {
			matches = append(matches, i)
		}
	}
	if len(matches) == 0 {
		for i, sprint := range sprintList {
			if strings.Contains(normalizeName(sprint.Name), normalized) {
				matches = append(matches, i)
			}
		}
	}
	if len(matches) != 1 {
		return sprints.CoreSprint{}, teams.CoreTeam{}, &AppRealtimeToolResponse{
			Success: false, Sprints: toRealtimeVoiceSprints(sprintList, indexTeamsByID(teamList), 10),
			Message: "Ask which sprint the user meant.",
		}, nil
	}
	sprint := sprintList[matches[0]]
	team := indexTeamsByID(teamList)[sprint.Team]
	return sprint, team, nil, nil
}

func (h *Handlers) resolveRealtimeObjective(ctx context.Context, workspaceID, userID uuid.UUID, name, teamName string) (objectives.CoreObjective, teams.CoreTeam, *AppRealtimeToolResponse, error) {
	teamList, err := h.teams.List(ctx, workspaceID, userID)
	if err != nil {
		return objectives.CoreObjective{}, teams.CoreTeam{}, nil, fmt.Errorf("list objective teams: %w", err)
	}
	filters := map[string]any{}
	if strings.TrimSpace(teamName) != "" {
		team := resolveRealtimeTeam(teamList, teamName)
		if team == nil {
			return objectives.CoreObjective{}, teams.CoreTeam{}, ptr(realtimeTeamClarification(teamList, "Ask which team's objective the user meant.")), nil
		}
		filters["team_id"] = team.ID
	}
	objectiveList, err := h.objectives.List(ctx, workspaceID, userID, filters)
	if err != nil {
		return objectives.CoreObjective{}, teams.CoreTeam{}, nil, fmt.Errorf("list objectives: %w", err)
	}
	normalized := normalizeName(name)
	matches := make([]int, 0, len(objectiveList))
	for i, objective := range objectiveList {
		if normalizeName(objective.Name) == normalized {
			matches = append(matches, i)
		}
	}
	if len(matches) == 0 {
		for i, objective := range objectiveList {
			if strings.Contains(normalizeName(objective.Name), normalized) {
				matches = append(matches, i)
			}
		}
	}
	if len(matches) != 1 {
		teamsByID := indexTeamsByID(teamList)
		voiceObjectives := make([]AppRealtimeVoiceObjective, 0, len(objectiveList))
		for _, objective := range objectiveList {
			voiceObjectives = append(voiceObjectives, toRealtimeVoiceObjective(objective, teamsByID))
		}
		return objectives.CoreObjective{}, teams.CoreTeam{}, &AppRealtimeToolResponse{
			Success: false, Objectives: voiceObjectives,
			Message: "Ask which objective the user meant.",
		}, nil
	}
	objective := objectiveList[matches[0]]
	return objective, indexTeamsByID(teamList)[objective.Team], nil, nil
}

func realtimeNavigationResponse(path, target string) AppRealtimeToolResponse {
	return AppRealtimeToolResponse{
		Success: true, Message: "Opening " + strings.ReplaceAll(target, "-", " ") + ".",
		ClientAction: &AppRealtimeClientAction{Type: "navigate", Path: path},
	}
}

func realtimeTeamClarification(teamList []teams.CoreTeam, message string) AppRealtimeToolResponse {
	return AppRealtimeToolResponse{
		Success: false, NeedsTeam: true, Teams: toRealtimeVoiceTeams(teamList), Message: message,
	}
}

func changedConfirmationResponse(token string) AppRealtimeToolResponse {
	return AppRealtimeToolResponse{
		Success: false, RequiresConfirmation: true, ConfirmationToken: token,
		Message: "The details changed after confirmation. Read back the current details and ask the user to confirm them again.",
	}
}

func decodeRealtimeArguments(rawArgs json.RawMessage, target any, toolName string) error {
	if len(rawArgs) == 0 {
		return nil
	}
	if err := json.Unmarshal(rawArgs, target); err != nil {
		return fmt.Errorf("invalid %s arguments: %w", toolName, err)
	}
	return nil
}

func resolveRealtimeStatus(statuses []states.CoreState, value string) *states.CoreState {
	normalized := normalizeName(value)
	for i := range statuses {
		if normalizeName(statuses[i].Name) == normalized || normalizeName(statuses[i].Category) == normalized {
			return &statuses[i]
		}
	}
	var match *states.CoreState
	for i := range statuses {
		if strings.Contains(normalizeName(statuses[i].Name), normalized) {
			if match != nil {
				return nil
			}
			match = &statuses[i]
		}
	}
	return match
}

func toRealtimeVoiceSprint(value sprints.CoreSprint, teamName string) AppRealtimeVoiceSprint {
	goal := ""
	if value.Goal != nil {
		goal = *value.Goal
	}
	completion := 0
	if value.TotalStories > 0 {
		completion = value.CompletedStories * 100 / value.TotalStories
	}
	return AppRealtimeVoiceSprint{
		Name: value.Name, Team: teamName, Goal: goal, StartDate: value.StartDate,
		EndDate: value.EndDate, TotalStories: value.TotalStories,
		CompletedStories: value.CompletedStories, StartedStories: value.StartedStories,
		CompletionPercentage: completion,
	}
}

func toRealtimeVoiceSprints(sprintList []sprints.CoreSprint, teamsByID map[uuid.UUID]teams.CoreTeam, limit int) []AppRealtimeVoiceSprint {
	result := make([]AppRealtimeVoiceSprint, 0, min(limit, len(sprintList)))
	for _, sprint := range sprintList {
		result = append(result, toRealtimeVoiceSprint(sprint, teamsByID[sprint.Team].Name))
		if len(result) >= limit {
			break
		}
	}
	return result
}

func feedbackItems(matches []realtimeFeedbackMatch) []AppRealtimeVoiceFeedbackItem {
	result := make([]AppRealtimeVoiceFeedbackItem, 0, len(matches))
	for _, match := range matches {
		result = append(result, match.item)
	}
	return result
}

func responseOrEmpty(response *AppRealtimeToolResponse) AppRealtimeToolResponse {
	if response == nil {
		return AppRealtimeToolResponse{}
	}
	return *response
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}

func displayWorkloadMemberName(fullName, username string) string {
	return firstNonEmpty(strings.TrimSpace(fullName), strings.TrimSpace(username), "Unknown member")
}

func activityPluralEnding(count int) string {
	if count == 1 {
		return "y"
	}
	return "ies"
}
