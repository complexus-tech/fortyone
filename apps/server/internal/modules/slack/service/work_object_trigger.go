package slack

import (
	"context"
	"errors"
	"strings"
	"time"

	slackdomain "github.com/complexus-tech/projects-api/internal/modules/slack/domain"
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
		if isSlackRepositoryNotFound(err) {
			return nil
		}
		return err
	}
	return s.processSlackEntityDetailsEvent(ctx, installation, event)
}

func (s *Service) processSlackEntityDetailsEvent(
	ctx context.Context,
	installation slackdomain.Installation,
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
	storyLink, storyLinkErr := ParseFortyOneStoryURL(event.EntityURL)
	requestLink, requestLinkErr := ParseFortyOneRequestURL(event.EntityURL)
	objectiveLink, objectiveLinkErr := ParseFortyOneObjectiveURL(event.EntityURL)
	sprintLink, sprintLinkErr := ParseFortyOneSprintURL(event.EntityURL)
	if storyLinkErr != nil && requestLinkErr != nil && objectiveLinkErr != nil && sprintLinkErr != nil {
		return nil
	}
	isRequest := requestLinkErr == nil
	isObjective := objectiveLinkErr == nil
	isSprint := sprintLinkErr == nil
	linkWorkspaceSlug := storyLink.WorkspaceSlug
	externalRefType := slackStoryExternalRefType
	if isRequest {
		linkWorkspaceSlug = requestLink.WorkspaceSlug
		externalRefType = slackRequestExternalRefType
	} else if isObjective {
		linkWorkspaceSlug = objectiveLink.WorkspaceSlug
		externalRefType = slackObjectiveExternalRefType
	} else if isSprint {
		linkWorkspaceSlug = sprintLink.WorkspaceSlug
		externalRefType = slackSprintExternalRefType
	}
	if !strings.EqualFold(linkWorkspaceSlug, workspace.Slug) {
		return nil
	}
	if event.ExternalRef.Type != "" && !strings.EqualFold(event.ExternalRef.Type, externalRefType) {
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
	if isRequest {
		return s.presentSlackRequestEntityDetails(ctx, publisher, botToken, workspace, installation, event, *linkedUserID, requestLink)
	}
	if isObjective {
		return s.presentSlackObjectiveEntityDetails(ctx, publisher, botToken, workspace, installation, event, *linkedUserID, objectiveLink)
	}
	if isSprint {
		return s.presentSlackSprintEntityDetails(ctx, publisher, botToken, workspace, installation, event, *linkedUserID, sprintLink)
	}

	storyReader, ok := s.stories.(SlackStoryReader)
	if !ok {
		return errors.New("slack Work Object story reader is not configured")
	}
	story, err := storyReader.QueryByRefForUser(ctx, workspace.ID, *linkedUserID, storyLink.StoryReference)
	if err != nil {
		if errors.Is(err, ErrStoryNotFound) || errors.Is(err, ErrInvalidStoryReference) || isSlackRepositoryNotFound(err) {
			return nil
		}
		return err
	}
	if !slackStoryExternalRefMatches(storyLink, story.ID.String(), event.ExternalRef.ID) {
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
		return errors.New("slack Work Object repository is not configured")
	}
	input, err := buildSlackStoryWorkObjectInput(ctx, repository, *linkedUserID, event.UserID, storyLink.CanonicalURL, story, true)
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

func (s *Service) presentSlackObjectiveEntityDetails(
	ctx context.Context,
	publisher *slackWorkObjectPublisher,
	botToken string,
	workspace slackdomain.Workspace,
	installation slackdomain.Installation,
	event normalizedSlackEvent,
	actorID uuid.UUID,
	link FortyOneObjectiveLink,
) error {
	if s.objectiveReader == nil {
		return nil
	}
	if !validSlackObjectiveExternalRef(link, event.ExternalRef.ID) {
		return nil
	}
	objectiveItems, err := s.objectiveReader.ListByID(ctx, workspace.ID, actorID, link.ObjectiveID)
	if err != nil {
		return err
	}
	var objective Objective
	found := false
	for _, candidate := range objectiveItems {
		if candidate.ID == link.ObjectiveID && candidate.Workspace == workspace.ID && candidate.Team == link.TeamID {
			objective = candidate
			found = true
			break
		}
	}
	if !found {
		return nil
	}
	source := requestSourceContext{
		SlackTeamID:    event.TeamID,
		SlackChannelID: event.ChannelID,
		SlackMessageTS: event.MessageTS,
		SlackThreadTS:  event.ThreadTS,
		SlackUserID:    event.UserID,
	}
	if err := s.ensureTeamAvailableForSlackSource(ctx, workspace.ID, actorID, objective.Team, source); err != nil {
		if errors.Is(err, ErrSlackTeamNotAvailable) {
			return nil
		}
		return err
	}
	repository, ok := s.repo.(slackWorkObjectRepository)
	if !ok {
		return errors.New("slack Work Object repository is not configured")
	}
	input, err := buildSlackObjectiveWorkObjectInput(ctx, repository, actorID, event.UserID, link.CanonicalURL, objective)
	if err != nil {
		return err
	}
	details, err := BuildSlackObjectiveEntityDetailsRequest(event.TriggerID, input)
	if err != nil {
		return err
	}
	if err := s.requireCurrentSlackInteractionActor(ctx, workspace.ID, actorID, source); err != nil {
		return err
	}
	if err := s.requireCurrentSlackInstallation(ctx, workspace.ID, event.TeamID, installation.InstallGeneration); err != nil {
		return err
	}
	return publisher.PresentDetails(ctx, botToken, details)
}

func (s *Service) presentSlackSprintEntityDetails(
	ctx context.Context,
	publisher *slackWorkObjectPublisher,
	botToken string,
	workspace slackdomain.Workspace,
	installation slackdomain.Installation,
	event normalizedSlackEvent,
	actorID uuid.UUID,
	link FortyOneSprintLink,
) error {
	if s.sprintReader == nil {
		return nil
	}
	if !validSlackSprintExternalRef(link, event.ExternalRef.ID) {
		return nil
	}
	sprintItems, err := s.sprintReader.ListByID(ctx, workspace.ID, actorID, link.SprintID)
	if err != nil {
		return err
	}
	var sprint Sprint
	found := false
	for _, candidate := range sprintItems {
		if candidate.ID == link.SprintID && candidate.WorkspaceID == workspace.ID && candidate.TeamID == link.TeamID {
			sprint = candidate
			found = true
			break
		}
	}
	if !found {
		return nil
	}
	source := requestSourceContext{
		SlackTeamID:    event.TeamID,
		SlackChannelID: event.ChannelID,
		SlackMessageTS: event.MessageTS,
		SlackThreadTS:  event.ThreadTS,
		SlackUserID:    event.UserID,
	}
	if err := s.ensureTeamAvailableForSlackSource(ctx, workspace.ID, actorID, sprint.TeamID, source); err != nil {
		if errors.Is(err, ErrSlackTeamNotAvailable) {
			return nil
		}
		return err
	}
	input := slackSprintWorkObjectInput(link.CanonicalURL, sprint)
	details, err := BuildSlackSprintEntityDetailsRequest(event.TriggerID, input)
	if err != nil {
		return err
	}
	if err := s.requireCurrentSlackInteractionActor(ctx, workspace.ID, actorID, source); err != nil {
		return err
	}
	if err := s.requireCurrentSlackInstallation(ctx, workspace.ID, event.TeamID, installation.InstallGeneration); err != nil {
		return err
	}
	return publisher.PresentDetails(ctx, botToken, details)
}

func (s *Service) presentSlackRequestEntityDetails(
	ctx context.Context,
	publisher *slackWorkObjectPublisher,
	botToken string,
	workspace slackdomain.Workspace,
	installation slackdomain.Installation,
	event normalizedSlackEvent,
	actorID uuid.UUID,
	link FortyOneRequestLink,
) error {
	request, err := s.requests.GetForUser(ctx, workspace.ID, link.RequestID, actorID)
	if err != nil {
		if isSlackRepositoryNotFound(err) {
			return nil
		}
		return err
	}
	if request.ID != link.RequestID || request.WorkspaceID != workspace.ID || request.TeamID != link.TeamID || !validSlackRequestExternalRef(link, event.ExternalRef.ID) {
		return nil
	}
	source := requestSourceContext{
		SlackTeamID:    event.TeamID,
		SlackChannelID: event.ChannelID,
		SlackMessageTS: event.MessageTS,
		SlackThreadTS:  event.ThreadTS,
		SlackUserID:    event.UserID,
	}
	if err := s.ensureTeamAvailableForSlackSource(ctx, workspace.ID, actorID, request.TeamID, source); err != nil {
		if errors.Is(err, ErrSlackTeamNotAvailable) {
			return nil
		}
		return err
	}
	repository, ok := s.repo.(slackWorkObjectRepository)
	if !ok {
		return errors.New("slack Work Object repository is not configured")
	}
	input, err := buildSlackRequestWorkObjectInput(ctx, repository, actorID, event.UserID, link.CanonicalURL, request)
	if err != nil {
		return err
	}
	details, err := BuildSlackRequestEntityDetailsRequest(event.TriggerID, input)
	if err != nil {
		return err
	}
	if err := s.requireCurrentSlackInteractionActor(ctx, workspace.ID, actorID, source); err != nil {
		return err
	}
	if err := s.ensureTeamAvailableForSlackSource(ctx, workspace.ID, actorID, request.TeamID, source); err != nil {
		if errors.Is(err, ErrSlackTeamNotAvailable) {
			return nil
		}
		return err
	}
	if err := s.requireCurrentSlackInstallation(ctx, workspace.ID, event.TeamID, installation.InstallGeneration); err != nil {
		return err
	}
	return publisher.PresentDetails(ctx, botToken, details)
}
