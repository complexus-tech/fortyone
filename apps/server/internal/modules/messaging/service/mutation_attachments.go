package messaging

import (
	"context"
	"encoding/json"
	"fmt"
	"html"
	"strings"

	"github.com/google/uuid"
)

const maximumConversationFiles = 50

func normalizedAttachmentSource(source StoryAttachmentSource) (StoryAttachmentSource, error) {
	source.Provider = strings.TrimSpace(source.Provider)
	source.ExternalWorkspaceID = strings.TrimSpace(source.ExternalWorkspaceID)
	source.ChannelID = strings.TrimSpace(source.ChannelID)
	source.ThreadTS = strings.TrimSpace(source.ThreadTS)
	source.MessageTS = strings.TrimSpace(source.MessageTS)
	source.FileID = strings.TrimSpace(source.FileID)
	source.Name = strings.TrimSpace(source.Name)
	if source.Provider != "slack" || source.ExternalWorkspaceID == "" || source.ChannelID == "" ||
		source.ThreadTS == "" || source.FileID == "" || source.Name == "" ||
		len(source.ExternalWorkspaceID) > 64 || len(source.ChannelID) > 64 || len(source.ThreadTS) > 32 || len(source.MessageTS) > 32 ||
		len(source.FileID) > 64 || len(source.Name) > 255 {
		return StoryAttachmentSource{}, fmt.Errorf("%w: invalid conversation file reference", ErrInvalidRequest)
	}
	for _, value := range []string{source.ExternalWorkspaceID, source.ChannelID, source.FileID} {
		if !validSlackFileReferenceID(value) {
			return StoryAttachmentSource{}, fmt.Errorf("%w: invalid conversation file identifier", ErrInvalidRequest)
		}
	}
	if !validSlackFileReferenceTS(source.ThreadTS) ||
		(source.MessageTS != "" && !validSlackFileReferenceTS(source.MessageTS)) {
		return StoryAttachmentSource{}, fmt.Errorf("%w: invalid conversation file timestamp", ErrInvalidRequest)
	}
	return source, nil
}

func validSlackFileReferenceID(value string) bool {
	if value == "" {
		return false
	}
	for _, char := range value {
		if (char < 'A' || char > 'Z') && (char < 'a' || char > 'z') &&
			(char < '0' || char > '9') && char != '-' && char != '_' {
			return false
		}
	}
	return true
}

func validSlackFileReferenceTS(value string) bool {
	if value == "" {
		return false
	}
	dotCount := 0
	for _, char := range value {
		if char == '.' {
			dotCount++
			continue
		}
		if char < '0' || char > '9' {
			return false
		}
	}
	return dotCount == 1 && value[0] != '.' && value[len(value)-1] != '.'
}

func normalizedAttachmentSources(sources []StoryAttachmentSource) ([]StoryAttachmentSource, error) {
	if len(sources) > maximumConversationFiles {
		return nil, fmt.Errorf("%w: too many conversation files", ErrInvalidRequest)
	}
	if len(sources) == 0 {
		return nil, nil
	}
	result := make([]StoryAttachmentSource, 0, len(sources))
	seen := make(map[string]struct{}, len(sources))
	for _, source := range sources {
		normalized, err := normalizedAttachmentSource(source)
		if err != nil {
			return nil, err
		}
		if _, exists := seen[normalized.FileID]; exists {
			continue
		}
		seen[normalized.FileID] = struct{}{}
		result = append(result, normalized)
	}
	return result, nil
}

func selectedConversationFile(scope ToolScope, rawFileID *string) (*StoryAttachmentSource, error) {
	if rawFileID == nil {
		return nil, nil
	}
	fileID := strings.TrimSpace(*rawFileID)
	if fileID == "" {
		return nil, fmt.Errorf("%w: file_id is required", ErrInvalidToolArguments)
	}
	for _, file := range scope.AvailableFiles {
		if file.FileID == fileID {
			copy := file
			return &copy, nil
		}
	}
	return nil, fmt.Errorf("%w: file_id is not in the current conversation", ErrInvalidToolArguments)
}

func (m *storyMutationExecutor) proposeAttachFile(
	ctx context.Context,
	executor *FortyOneToolExecutor,
	scope ToolScope,
	raw json.RawMessage,
) (json.RawMessage, error) {
	var args struct {
		StoryID        *string `json:"story_id"`
		StoryReference *string `json:"story_reference"`
		FileID         string  `json:"file_id"`
	}
	if err := decodeToolArguments(raw, &args, "story_id", "story_reference", "file_id"); err != nil {
		return nil, err
	}
	if (args.StoryID == nil) == (args.StoryReference == nil) {
		return nil, fmt.Errorf("%w: provide exactly one of story_id or story_reference", ErrInvalidToolArguments)
	}
	file, err := selectedConversationFile(scope, &args.FileID)
	if err != nil {
		return nil, err
	}
	_, joinedByID, err := executor.joinedTeams(ctx, scope)
	if err != nil {
		return nil, err
	}
	story, err := m.resolveUpdateStory(ctx, scope, joinedByID, args.StoryID, args.StoryReference)
	if err != nil {
		return nil, err
	}
	team, allowed := joinedByID[story.Team]
	if !allowed || story.Workspace != scope.WorkspaceID {
		return nil, fmt.Errorf("%w: %s", ErrTeamNotAccessible, story.Team)
	}
	confirmationID, err := uuid.NewRandomFromReader(m.random)
	if err != nil {
		return nil, fmt.Errorf("generate story mutation confirmation ID: %w", err)
	}
	claims := storyMutationClaims{
		Version:        storyMutationTokenVersion,
		ConfirmationID: confirmationID,
		Operation:      StoryMutationAttachFile,
		WorkspaceID:    scope.WorkspaceID,
		UserID:         scope.UserID,
		TeamID:         story.Team,
		StoryID:        &story.ID,
		Attachment:     file,
		ExpiresAt:      m.now().UTC().Add(storyMutationConfirmationTTL),
	}
	reference := storyReference(team.Code, story.SequenceID)
	return m.marshalProposal(ctx, claims, StoryMutationPreview{
		StoryID: &story.ID, Reference: reference, TeamID: team.ID,
		TeamName: team.Name, TeamCode: strings.ToUpper(team.Code), Title: story.Title,
		ChangedFields: []string{"attachment"},
	}, fmt.Sprintf("Attach %q to %s?", html.EscapeString(file.Name), reference))
}

func (m *storyMutationExecutor) confirmAttachFile(
	ctx context.Context,
	scope ToolScope,
	team messagingTeam,
	claims storyMutationClaims,
) (StoryMutationResult, error) {
	if claims.StoryID == nil || claims.Attachment == nil {
		return StoryMutationResult{}, fmt.Errorf("%w: malformed file attachment proposal", ErrInvalidConfirmation)
	}
	story, err := m.stories.Get(ctx, *claims.StoryID, scope.WorkspaceID)
	if err != nil {
		return StoryMutationResult{}, fmt.Errorf("load story for confirmed attachment: %w", err)
	}
	if story.Team != team.ID || story.Workspace != scope.WorkspaceID {
		return StoryMutationResult{}, fmt.Errorf("%w: story team does not match proposal", ErrInvalidConfirmation)
	}
	result := storyMutationResult(storyMutationStatusApplied, StoryMutationAttachFile, story, team.Code)
	attachment := *claims.Attachment
	result.Attachment = &attachment
	return result, nil
}
