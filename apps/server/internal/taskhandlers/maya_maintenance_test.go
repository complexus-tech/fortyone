package taskhandlers

import (
	"context"
	"testing"
	"time"

	mayadomain "github.com/complexus-tech/projects-api/internal/modules/maya/domain"
	"github.com/hibiken/asynq"
	"github.com/stretchr/testify/require"
)

func TestMayaMaintenanceHandlersDelegateToTypedStore(t *testing.T) {
	store := &mayaWorkFocusStoreStub{}
	handlers := NewMayaMaintenanceHandlers(MayaMaintenanceHandlerDependencies{
		Log:   testTaskLogger(),
		Store: store,
	})

	require.NoError(t, handlers.HandleWorkFocusInference(t.Context(), asynq.NewTask("test", nil)))
	require.Equal(t, 1, store.listCandidatesCalls)
}

func TestMayaMaintenanceHandlersRequireCoreDependencies(t *testing.T) {
	var handlers *MayaMaintenanceHandlers
	err := handlers.HandleWorkFocusInference(t.Context(), nil)
	require.ErrorContains(t, err, "dependencies are required")

	handlers = NewMayaMaintenanceHandlers(MayaMaintenanceHandlerDependencies{Log: testTaskLogger()})
	err = handlers.HandleWorkFocusInference(t.Context(), nil)
	require.ErrorContains(t, err, "maya work focus store is not configured")
}

type mayaWorkFocusStoreStub struct {
	listCandidatesCalls int
}

func (store *mayaWorkFocusStoreStub) ListMayaWorkFocusCandidates(
	context.Context,
	int,
) ([]mayadomain.WorkFocusMember, error) {
	store.listCandidatesCalls++
	return nil, nil
}

func (*mayaWorkFocusStoreStub) ListMayaWorkFocusEvidence(
	context.Context,
	mayadomain.WorkFocusMember,
	time.Time,
	int,
) ([]mayadomain.WorkFocusEvidence, error) {
	return nil, nil
}

func (*mayaWorkFocusStoreStub) SaveMayaInferredWorkFocus(
	context.Context,
	mayadomain.WorkFocusMember,
	mayadomain.WorkFocusInferenceResult,
) (bool, error) {
	return false, nil
}
