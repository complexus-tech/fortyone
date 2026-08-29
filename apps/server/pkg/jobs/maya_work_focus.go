package jobs

import (
	"context"
	"fmt"
	"time"

	mayadomain "github.com/complexus-tech/projects-api/internal/modules/maya/domain"
	maya "github.com/complexus-tech/projects-api/internal/modules/maya/service"
	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/complexus-tech/projects-api/pkg/web"
)

const (
	mayaWorkFocusMemberBatchSize = 500
	mayaWorkFocusEvidenceLimit   = 30
	mayaWorkFocusLookback        = 180 * 24 * time.Hour
)

// MayaWorkFocusStore is the job-owned persistence contract. The Maya
// repository implements it with generated SQLC methods; the policy below stays
// deterministic and independent from a database driver.
type MayaWorkFocusStore interface {
	ListMayaWorkFocusCandidates(context.Context, int) ([]mayadomain.WorkFocusMember, error)
	ListMayaWorkFocusEvidence(
		context.Context,
		mayadomain.WorkFocusMember,
		time.Time,
		int,
	) ([]mayadomain.WorkFocusEvidence, error)
	SaveMayaInferredWorkFocus(
		context.Context,
		mayadomain.WorkFocusMember,
		mayadomain.WorkFocusInferenceResult,
	) (bool, error)
}

func ProcessMayaWorkFocusInference(
	ctx context.Context,
	store MayaWorkFocusStore,
	log *logger.Logger,
	asOf time.Time,
) error {
	ctx, span := web.AddSpan(ctx, "jobs.ProcessMayaWorkFocusInference")
	defer span.End()

	if store == nil {
		return fmt.Errorf("maya work focus store is not configured")
	}
	if asOf.IsZero() {
		return fmt.Errorf("maya work focus inference time is required")
	}
	asOf = asOf.UTC()

	members, err := store.ListMayaWorkFocusCandidates(ctx, mayaWorkFocusMemberBatchSize)
	if err != nil {
		return fmt.Errorf("list Maya work focus candidates: %w", err)
	}
	if len(members) == 0 {
		if log != nil {
			log.Info(ctx, "Maya work focus inference skipped: no candidates")
		}
		return nil
	}

	inferredCount := 0
	for _, member := range members {
		evidence, err := store.ListMayaWorkFocusEvidence(
			ctx,
			member,
			asOf.Add(-mayaWorkFocusLookback),
			mayaWorkFocusEvidenceLimit,
		)
		if err != nil {
			if log != nil {
				log.Error(ctx, "failed to load Maya work focus evidence", "workspace_id", member.WorkspaceID, "team_id", member.TeamID, "user_id", member.UserID, "error", err)
			}
			continue
		}
		result := maya.InferWorkFocus(maya.WorkFocusInferenceInput{
			ManualRoleTitle:       member.ManualRoleTitle,
			ManualRoleDescription: member.ManualRoleDescription,
			Evidence:              evidence,
		})
		if !result.ShouldInfer {
			continue
		}
		updated, err := store.SaveMayaInferredWorkFocus(ctx, member, result)
		if err != nil {
			if log != nil {
				log.Error(ctx, "failed to save Maya inferred work focus", "workspace_id", member.WorkspaceID, "team_id", member.TeamID, "user_id", member.UserID, "error", err)
			}
			continue
		}
		if updated {
			inferredCount++
		}
	}

	if log != nil {
		log.Info(ctx, "Maya work focus inference completed", "candidate_count", len(members), "inferred_count", inferredCount)
	}
	return nil
}
