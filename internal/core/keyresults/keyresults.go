package keyresults

import (
	"context"
	"errors"
	"fmt"

	"github.com/complexus-tech/projects-api/internal/repo/keyresultsrepo"
	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/complexus-tech/projects-api/pkg/web"
	"github.com/google/uuid"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// Set of error variables for key result operations
var (
	ErrNotFound = errors.New("key result not found")
)

// Service manages the key result operations
type Service struct {
	repo keyresultsrepo.Repository
	log  *logger.Logger
}

// New creates a new key result service
func New(log *logger.Logger, repo keyresultsrepo.Repository) *Service {
	return &Service{
		repo: repo,
		log:  log,
	}
}

// Create inserts a new key result into the system
func (s *Service) Create(ctx context.Context, nkr CoreNewKeyResult) (CoreKeyResult, error) {

	kr := keyresultsrepo.CoreKeyResult{
		ObjectiveID:     nkr.ObjectiveID,
		Name:            nkr.Name,
		MeasurementType: nkr.MeasurementType,
		StartValue:      nkr.StartValue,
		CurrentValue:    nkr.CurrentValue,
		TargetValue:     nkr.TargetValue,
		Lead:            nkr.Lead,
		Contributors:    nkr.Contributors,
		StartDate:       nkr.StartDate,
		EndDate:         nkr.EndDate,
		CreatedBy:       nkr.CreatedBy,
	}

	if err := s.repo.Create(ctx, &kr); err != nil {
		return CoreKeyResult{}, fmt.Errorf("create: %w", err)
	}

	// Add contributors if provided
	if len(nkr.Contributors) > 0 {
		if err := s.repo.AddContributors(ctx, kr.ID, nkr.Contributors); err != nil {
			return CoreKeyResult{}, fmt.Errorf("failed to add contributors: %w", err)
		}
	}

	return CoreKeyResult{
		ID:              kr.ID,
		ObjectiveID:     kr.ObjectiveID,
		Name:            kr.Name,
		MeasurementType: kr.MeasurementType,
		StartValue:      kr.StartValue,
		CurrentValue:    kr.CurrentValue,
		TargetValue:     kr.TargetValue,
		Lead:            kr.Lead,
		Contributors:    kr.Contributors,
		StartDate:       kr.StartDate,
		EndDate:         kr.EndDate,
		CreatedAt:       kr.CreatedAt,
		UpdatedAt:       kr.UpdatedAt,
		CreatedBy:       kr.CreatedBy,
	}, nil
}

// Update updates a key result in the system
func (s *Service) Update(ctx context.Context, id uuid.UUID, workspaceId uuid.UUID, updates map[string]any) error {
	ctx, span := web.AddSpan(ctx, "business.core.keyresults.Update")
	defer span.End()

	// Handle contributors separately
	var contributors []uuid.UUID
	if contribVal, exists := updates["contributors"]; exists {
		if contribSlice, ok := contribVal.([]uuid.UUID); ok {
			contributors = contribSlice
		}
		delete(updates, "contributors") // Remove from updates map
	}

	if err := s.repo.Update(ctx, id, workspaceId, updates); err != nil {
		if errors.Is(err, keyresultsrepo.ErrNotFound) {
			return ErrNotFound
		}
		return err
	}

	// Update contributors if provided
	if contributors != nil {
		if err := s.repo.UpdateContributors(ctx, id, contributors); err != nil {
			return fmt.Errorf("failed to update contributors: %w", err)
		}
	}

	span.AddEvent("key result updated", trace.WithAttributes(
		attribute.String("key_result.id", id.String()),
	))

	return nil
}

// Delete removes a key result from the system
func (s *Service) Delete(ctx context.Context, id uuid.UUID, workspaceId uuid.UUID) error {
	ctx, span := web.AddSpan(ctx, "business.core.keyresults.Delete")
	defer span.End()

	if err := s.repo.Delete(ctx, id, workspaceId); err != nil {
		if errors.Is(err, keyresultsrepo.ErrNotFound) {
			return ErrNotFound
		}
		return err
	}

	span.AddEvent("key result deleted", trace.WithAttributes(
		attribute.String("key_result.id", id.String()),
	))

	return nil
}

// Get retrieves a key result from the system
func (s *Service) Get(ctx context.Context, id uuid.UUID, workspaceId uuid.UUID) (CoreKeyResult, error) {
	ctx, span := web.AddSpan(ctx, "business.core.keyresults.Get")
	defer span.End()

	kr, err := s.repo.Get(ctx, id, workspaceId)
	if err != nil {
		if errors.Is(err, keyresultsrepo.ErrNotFound) {
			return CoreKeyResult{}, ErrNotFound
		}
		return CoreKeyResult{}, err
	}

	return CoreKeyResult(kr), nil
}

// List retrieves all key results for an objective
func (s *Service) List(ctx context.Context, objectiveId uuid.UUID, workspaceId uuid.UUID) ([]CoreKeyResult, error) {
	ctx, span := web.AddSpan(ctx, "business.core.keyresults.List")
	defer span.End()

	krs, err := s.repo.List(ctx, objectiveId, workspaceId)
	if err != nil {
		return nil, err
	}

	results := make([]CoreKeyResult, len(krs))
	for i, kr := range krs {
		results[i] = CoreKeyResult(kr)
	}

	return results, nil
}

// ListPaginated retrieves paginated key results with filters
func (s *Service) ListPaginated(ctx context.Context, filters keyresultsrepo.CoreKeyResultFilters) (keyresultsrepo.CoreKeyResultListResponse, error) {
	ctx, span := web.AddSpan(ctx, "business.core.keyresults.ListPaginated")
	defer span.End()

	response, err := s.repo.ListPaginated(ctx, filters)
	if err != nil {
		return keyresultsrepo.CoreKeyResultListResponse{}, fmt.Errorf("listing paginated key results: %w", err)
	}

	return response, nil
}
