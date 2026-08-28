package jobs

import (
	"context"
	"testing"
	"time"

	feedback "github.com/complexus-tech/projects-api/internal/modules/feedback/service"
	"github.com/stretchr/testify/require"
)

func TestPurgeDeletedFeedbackDelegatesRetentionToFeedbackStore(t *testing.T) {
	store := &recordingFeedbackMaintenanceStore{
		deletedResult: feedback.CoreDeletedFeedbackPurgeResult{ItemsDeleted: 3, ContributorsDeleted: 2},
	}

	err := PurgeDeletedFeedback(context.Background(), store, newTestJobLogger())

	require.NoError(t, err)
	require.WithinDuration(t, time.Now().UTC().Add(-30*24*time.Hour), store.deletedBefore, time.Second)
}

type recordingFeedbackMaintenanceStore struct {
	cutoffs        feedback.CoreContributorArtifactCutoffs
	artifactResult feedback.CoreContributorArtifactPurgeResult
	deletedBefore  time.Time
	deletedResult  feedback.CoreDeletedFeedbackPurgeResult
}

func (s *recordingFeedbackMaintenanceStore) PurgeExpiredContributorArtifacts(_ context.Context, cutoffs feedback.CoreContributorArtifactCutoffs) (feedback.CoreContributorArtifactPurgeResult, error) {
	s.cutoffs = cutoffs
	return s.artifactResult, nil
}

func (s *recordingFeedbackMaintenanceStore) PurgeDeletedFeedback(_ context.Context, deletedBefore time.Time) (feedback.CoreDeletedFeedbackPurgeResult, error) {
	s.deletedBefore = deletedBefore
	return s.deletedResult, nil
}
