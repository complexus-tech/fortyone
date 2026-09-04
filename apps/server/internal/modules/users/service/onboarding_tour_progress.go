package users

import (
	"context"
	"fmt"

	apptracing "github.com/complexus-tech/projects-api/pkg/tracing"
	"github.com/google/uuid"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// GetOnboardingTourProgress retrieves (or creates) the versioned onboarding
// progress for an active user.
func (s *Service) GetOnboardingTourProgress(
	ctx context.Context,
	userID uuid.UUID,
	scope CoreOnboardingTourScope,
) (CoreOnboardingTourProgress, error) {
	s.log.Info(ctx, "business.core.users.GetOnboardingTourProgress")
	ctx, span := apptracing.AddSpanFromContext(ctx, "business.core.users.GetOnboardingTourProgress")
	defer span.End()

	if err := scope.NormalizeAndValidate(); err != nil {
		return CoreOnboardingTourProgress{}, err
	}
	progress, err := s.repo.GetOnboardingTourProgress(ctx, userID, scope)
	if err != nil {
		span.RecordError(err)
		return CoreOnboardingTourProgress{}, fmt.Errorf("get onboarding tour progress: %w", err)
	}

	span.AddEvent("onboarding tour progress retrieved", trace.WithAttributes(
		attribute.String("user.id", userID.String()),
	))
	return progress, nil
}

// UpdateOnboardingTourProgress merges newly resolved steps/actions into the
// versioned progress record for an active user.
func (s *Service) UpdateOnboardingTourProgress(
	ctx context.Context,
	userID uuid.UUID,
	updates CoreUpdateOnboardingTourProgress,
) (CoreOnboardingTourProgress, error) {
	s.log.Info(ctx, "business.core.users.UpdateOnboardingTourProgress")
	ctx, span := apptracing.AddSpanFromContext(ctx, "business.core.users.UpdateOnboardingTourProgress")
	defer span.End()

	if err := updates.NormalizeAndValidate(); err != nil {
		return CoreOnboardingTourProgress{}, err
	}
	progress, err := s.repo.UpdateOnboardingTourProgress(ctx, userID, updates)
	if err != nil {
		span.RecordError(err)
		return CoreOnboardingTourProgress{}, fmt.Errorf("update onboarding tour progress: %w", err)
	}

	span.AddEvent("onboarding tour progress updated", trace.WithAttributes(
		attribute.String("user.id", userID.String()),
	))
	return progress, nil
}
