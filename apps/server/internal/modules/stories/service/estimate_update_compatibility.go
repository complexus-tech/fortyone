package stories

import (
	"context"
	"fmt"

	"github.com/google/uuid"
)

func (s *Service) applyEstimateUpdate(ctx context.Context, workspaceID uuid.UUID, story CoreSingleStory, updates map[string]any) error {
	estimateRaw, hasEstimateUpdate := updates["estimate_unit"]
	if !hasEstimateUpdate {
		return nil
	}

	estimateScheme := story.EstimateScheme
	if estimateScheme == "" {
		var err error
		estimateScheme, err = s.getTeamEstimateScheme(ctx, workspaceID, story.Team)
		if err != nil {
			return err
		}
	}

	estimateValue, err := normalizeEstimateUpdateValue(estimateRaw)
	if err != nil {
		return err
	}

	if err := ValidateEstimateValue(estimateScheme, estimateValue); err != nil {
		if estimateValue != nil {
			return fmt.Errorf("%w. If this work is larger than the max estimate, split it into smaller stories", err)
		}
		return err
	}

	updates["estimate_unit"] = estimateValue
	return nil
}
