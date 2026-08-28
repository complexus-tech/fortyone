package reports

import (
	"context"
	"fmt"
	"strings"
	"time"

	apptracing "github.com/complexus-tech/projects-api/pkg/tracing"
	"github.com/google/uuid"
)

const (
	maxAnalyticsEventNameLength = 120
	maxAnalyticsSurfaceLength   = 80
)

func (s *Service) TrackWorkspaceAnalyticsEvent(ctx context.Context, input CoreWorkspaceAnalyticsEventInput) (CoreWorkspaceAnalyticsEventInput, error) {
	ctx, span := apptracing.AddSpanFromContext(ctx, "business.core.reports.TrackWorkspaceAnalyticsEvent")
	defer span.End()

	actorID, err := reportActorID(ctx, input.WorkspaceID)
	if err != nil || actorID != input.UserID {
		span.RecordError(ErrReportsAccessDenied)
		return CoreWorkspaceAnalyticsEventInput{}, ErrReportsAccessDenied
	}

	normalized, err := normalizeWorkspaceAnalyticsEventInput(input)
	if err != nil {
		span.RecordError(err)
		return CoreWorkspaceAnalyticsEventInput{}, err
	}

	if err := s.repo.CreateWorkspaceAnalyticsEvent(ctx, normalized); err != nil {
		span.RecordError(err)
		return CoreWorkspaceAnalyticsEventInput{}, fmt.Errorf("creating workspace analytics event: %w", err)
	}

	return normalized, nil
}

func normalizeWorkspaceAnalyticsEventInput(input CoreWorkspaceAnalyticsEventInput) (CoreWorkspaceAnalyticsEventInput, error) {
	if input.WorkspaceID == uuid.Nil || input.UserID == uuid.Nil {
		return CoreWorkspaceAnalyticsEventInput{}, ErrInvalidWorkspaceAnalyticsEvent
	}

	eventName := strings.TrimSpace(input.EventName)
	if eventName == "" || len(eventName) > maxAnalyticsEventNameLength {
		return CoreWorkspaceAnalyticsEventInput{}, ErrInvalidWorkspaceAnalyticsEvent
	}

	surface := strings.TrimSpace(input.Surface)
	if surface == "" {
		surface = "unknown"
	}
	if len(surface) > maxAnalyticsSurfaceLength {
		return CoreWorkspaceAnalyticsEventInput{}, ErrInvalidWorkspaceAnalyticsEvent
	}

	input.EventName = eventName
	input.Surface = surface
	if input.Properties == nil {
		input.Properties = map[string]any{}
	}
	if input.OccurredAt.IsZero() {
		input.OccurredAt = time.Now().UTC()
	} else {
		input.OccurredAt = input.OccurredAt.UTC()
	}

	return input, nil
}
