package slack

import (
	"context"
	"errors"
	"strings"
	"time"

	slackrepository "github.com/complexus-tech/projects-api/internal/modules/slack/repository"
	stories "github.com/complexus-tech/projects-api/internal/modules/stories/service"
	"github.com/google/uuid"
)

func (s *Service) newSlackWorkObjectTriggerContext(parent context.Context) (context.Context, context.CancelFunc) {
	timeout := s.workObjectTriggerTimeout
	if timeout <= 0 || timeout >= 3*time.Second {
		timeout = slackWorkObjectTriggerTimeout
	}
	return context.WithTimeout(parent, timeout)
}

func (s *Service) requireCurrentSlackInteractionActor(
	ctx context.Context,
	workspaceID, expectedActorID uuid.UUID,
	source requestSourceContext,
) error {
	actorID, err := s.findLinkedInteractionActor(ctx, workspaceID, source)
	if err != nil {
		return err
	}
	if actorID != expectedActorID {
		return ErrSlackInteractionActorMismatch
	}
	return nil
}

// handleSlackEntityDetailsEvent owns the latency-sensitive Work Object path.
// The caller must give it a context that expires before Slack's trigger does
// and must never hand this event to the durable inbox after this method returns.
func (s *Service) handleSlackEntityDetailsEvent(ctx context.Context, event normalizedSlackEvent) error {
	installation, err := s.repo.GetSlackWorkspaceByTeamID(ctx, event.TeamID)
	if err != nil {
		if slackrepository.IsNotFound(err) {
			return nil
		}
		return err
	}
	return s.processSlackEntityDetailsEvent(ctx, installation, event)
}

func (s *Service) processSlackEntityDetailsEvent(
	ctx context.Context,
	installation slackrepository.SlackWorkspaceRecord,
	event normalizedSlackEvent,
) error {
	if event.Kind != slackEventKindEntityDetails || !installation.IsActive || installation.ID == uuid.Nil || installation.WorkspaceID == uuid.Nil {
		return nil
	}
	if !strings.EqualFold(strings.TrimSpace(installation.SlackTeamID), event.TeamID) {
		return nil
	}

	workspace, err := s.repo.FindWorkspaceByID(ctx, installation.WorkspaceID)
	if err != nil {
		return err
	}
	link, err := ParseFortyOneStoryURL(event.EntityURL)
	if err != nil || !strings.EqualFold(link.WorkspaceSlug, workspace.Slug) {
		return nil
	}
	if event.ExternalRef.Type != "" && !strings.EqualFold(event.ExternalRef.Type, slackStoryExternalRefType) {
		return nil
	}

	linkedUserID, err := s.repo.FindLinkedUserIDBySlackUser(ctx, workspace.ID, event.TeamID, event.UserID)
	if err != nil {
		return err
	}
	botToken, err := s.botToken(ctx, installation)
	if err != nil {
		return err
	}
	publisher := newSlackWorkObjectPublisher(s.webClient)
	if linkedUserID == nil || *linkedUserID == uuid.Nil {
		authURL, linkErr := s.buildSlackUserLinkURL(ctx, workspace.ID, event.TeamID, event.UserID)
		if linkErr != nil {
			return linkErr
		}
		request, buildErr := BuildSlackStoryAuthenticationEntityDetailsRequest(event.TriggerID, authURL)
		if buildErr != nil {
			return buildErr
		}
		if currentErr := s.requireCurrentSlackInstallation(ctx, workspace.ID, event.TeamID, installation.InstallGeneration); currentErr != nil {
			return currentErr
		}
		return publisher.PresentDetails(ctx, botToken, request)
	}

	storyReader, ok := s.stories.(SlackStoryReader)
	if !ok {
		return errors.New("Slack Work Object story reader is not configured")
	}
	story, err := storyReader.QueryByRef(ctx, workspace.ID, link.StoryReference)
	if err != nil {
		if errors.Is(err, stories.ErrNotFound) || errors.Is(err, stories.ErrInvalidStoryReference) || slackrepository.IsNotFound(err) {
			return nil
		}
		return err
	}
	if !slackStoryExternalRefMatches(link, story.ID.String(), event.ExternalRef.ID) {
		return nil
	}
	source := requestSourceContext{
		SlackTeamID:    event.TeamID,
		SlackChannelID: event.ChannelID,
		SlackMessageTS: event.MessageTS,
		SlackThreadTS:  event.ThreadTS,
		SlackUserID:    event.UserID,
	}
	if err := s.ensureTeamAvailableForSlackSource(ctx, workspace.ID, *linkedUserID, story.Team, source); err != nil {
		if errors.Is(err, ErrSlackTeamNotAvailable) {
			return nil
		}
		return err
	}
	repository, ok := s.repo.(slackWorkObjectRepository)
	if !ok {
		return errors.New("Slack Work Object repository is not configured")
	}
	input, err := buildSlackStoryWorkObjectInput(ctx, repository, *linkedUserID, event.UserID, link.CanonicalURL, story, true)
	if err != nil {
		return err
	}
	request, err := BuildSlackStoryEntityDetailsRequest(event.TriggerID, input)
	if err != nil {
		return err
	}
	if err := s.requireCurrentSlackInteractionActor(ctx, workspace.ID, *linkedUserID, source); err != nil {
		return err
	}
	if err := s.ensureTeamAvailableForSlackSource(ctx, workspace.ID, *linkedUserID, story.Team, source); err != nil {
		if errors.Is(err, ErrSlackTeamNotAvailable) {
			return nil
		}
		return err
	}
	if err := s.requireCurrentSlackInstallation(ctx, workspace.ID, event.TeamID, installation.InstallGeneration); err != nil {
		return err
	}
	return publisher.PresentDetails(ctx, botToken, request)
}
