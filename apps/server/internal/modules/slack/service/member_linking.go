package slack

import (
	"context"
	"net/url"
	"strings"

	"github.com/google/uuid"
)

func (s *Service) autoLinkWorkspaceMembers(ctx context.Context, slackWorkspace slackWorkspaceRecord) error {
	if !slackInstallationHasScopes(slackWorkspace, "users:read", "users:read.email") {
		return ErrSlackMemberLinkingScopesMissing
	}
	botToken, err := s.botToken(ctx, slackWorkspace)
	if err != nil {
		return err
	}
	slackUsers, err := s.fetchWorkspaceUsers(ctx, botToken)
	if err != nil {
		return err
	}
	if len(slackUsers) == 0 {
		return nil
	}

	workspaceMembers, err := s.repo.ListWorkspaceMembersForSlackLinking(ctx, slackWorkspace.WorkspaceID)
	if err != nil {
		return err
	}
	if len(workspaceMembers) == 0 {
		return nil
	}

	memberByEmail := make(map[string]uuid.UUID, len(workspaceMembers))
	for _, member := range workspaceMembers {
		normalizedEmail := normalizeEmail(member.Email)
		if normalizedEmail == "" {
			continue
		}
		memberByEmail[normalizedEmail] = member.UserID
	}
	if len(memberByEmail) == 0 {
		return nil
	}

	links := make([]slackUserLinkUpsert, 0, len(slackUsers))
	for _, slackUser := range slackUsers {
		normalizedEmail := normalizeEmail(slackUser.Email)
		if normalizedEmail == "" {
			continue
		}
		userID, ok := memberByEmail[normalizedEmail]
		if !ok || userID == uuid.Nil {
			continue
		}
		links = append(links, slackUserLinkUpsert{
			SlackUserID: slackUser.ID,
			UserID:      userID,
			LinkedVia:   "email_match",
		})
	}
	if len(links) == 0 {
		return nil
	}

	return s.repo.UpsertSlackUserLinks(ctx, slackWorkspace.WorkspaceID, slackWorkspace.ID, slackWorkspace.SlackTeamID, links)
}

func slackInstallationHasScopes(slackWorkspace slackWorkspaceRecord, required ...string) bool {
	if slackWorkspace.Scope == nil {
		return false
	}
	available := make(map[string]struct{})
	for _, scope := range strings.FieldsFunc(*slackWorkspace.Scope, func(r rune) bool {
		return r == ',' || r == ' '
	}) {
		if scope = strings.TrimSpace(scope); scope != "" {
			available[scope] = struct{}{}
		}
	}
	for _, scope := range required {
		if _, ok := available[scope]; !ok {
			return false
		}
	}
	return true
}

func (s *Service) fetchWorkspaceUsers(ctx context.Context, botToken string) ([]slackWorkspaceUser, error) {
	cursor := ""
	users := make([]slackWorkspaceUser, 0)

	for {
		endpoint := "https://slack.com/api/users.list?limit=200"
		if cursor != "" {
			endpoint += "&cursor=" + url.QueryEscape(cursor)
		}
		var response struct {
			OK      bool   `json:"ok"`
			Error   string `json:"error"`
			Members []struct {
				ID       string `json:"id"`
				Name     string `json:"name"`
				RealName string `json:"real_name"`
				Deleted  bool   `json:"deleted"`
				IsBot    bool   `json:"is_bot"`
				Profile  struct {
					Email string `json:"email"`
				} `json:"profile"`
			} `json:"members"`
			ResponseMetadata struct {
				NextCursor string `json:"next_cursor"`
			} `json:"response_metadata"`
		}
		if err := s.callSlackAPI(ctx, botToken, endpoint, nil, &response); err != nil {
			return nil, err
		}

		for _, member := range response.Members {
			if member.Deleted || member.IsBot {
				continue
			}
			memberID := strings.TrimSpace(member.ID)
			if memberID == "" {
				continue
			}
			users = append(users, slackWorkspaceUser{
				ID:       memberID,
				Username: strings.TrimSpace(member.Name),
				FullName: strings.TrimSpace(member.RealName),
				Email:    strings.TrimSpace(member.Profile.Email),
			})
		}

		cursor = strings.TrimSpace(response.ResponseMetadata.NextCursor)
		if cursor == "" {
			break
		}
	}

	return users, nil
}

func (s *Service) resolveLinkedSlackUser(ctx context.Context, workspaceID uuid.UUID, source requestSourceContext) (uuid.UUID, string, error) {
	slackTeamID := strings.TrimSpace(source.SlackTeamID)
	slackUserID := strings.TrimSpace(source.SlackUserID)
	if slackTeamID != "" && slackUserID != "" {
		mappedUserID, err := s.repo.FindLinkedUserIDBySlackUser(ctx, workspaceID, slackTeamID, slackUserID)
		if err != nil {
			return uuid.Nil, "", err
		}
		if mappedUserID != nil && *mappedUserID != uuid.Nil {
			return *mappedUserID, "", nil
		}
	}

	connectURL, err := s.buildSlackUserLinkURL(ctx, workspaceID, slackTeamID, slackUserID)
	if err != nil {
		return uuid.Nil, "", err
	}
	return uuid.Nil, connectURL, nil
}

func interactionSourceForPayload(payload interactionPayload, source requestSourceContext) (requestSourceContext, error) {
	slackTeamID := strings.TrimSpace(payload.Team.ID)
	slackUserID := strings.TrimSpace(payload.User.ID)
	if slackTeamID == "" || slackUserID == "" {
		return requestSourceContext{}, ErrSlackInteractionActorMismatch
	}
	if sourceTeamID := strings.TrimSpace(source.SlackTeamID); sourceTeamID != "" && sourceTeamID != slackTeamID {
		return requestSourceContext{}, ErrSlackInteractionActorMismatch
	}
	if sourceUserID := strings.TrimSpace(source.SlackUserID); sourceUserID != "" && sourceUserID != slackUserID {
		return requestSourceContext{}, ErrSlackInteractionActorMismatch
	}

	source.SlackTeamID = slackTeamID
	source.SlackUserID = slackUserID
	if username := strings.TrimSpace(payload.User.Username); username != "" {
		source.SlackUsername = username
	} else if username := strings.TrimSpace(payload.User.Name); username != "" {
		source.SlackUsername = username
	}
	return source, nil
}

func (s *Service) findLinkedInteractionActor(ctx context.Context, workspaceID uuid.UUID, source requestSourceContext) (uuid.UUID, error) {
	linkedUserID, err := s.repo.FindLinkedUserIDBySlackUser(
		ctx,
		workspaceID,
		strings.TrimSpace(source.SlackTeamID),
		strings.TrimSpace(source.SlackUserID),
	)
	if err != nil {
		return uuid.Nil, err
	}
	if linkedUserID == nil || *linkedUserID == uuid.Nil {
		return uuid.Nil, ErrSlackUserNotLinked
	}
	return *linkedUserID, nil
}

type slackWorkspaceUser struct {
	ID       string
	Username string
	FullName string
	Email    string
}
