package emailreply

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestProcessorStateContextDetachesCancellationButKeepsShortDeadline(t *testing.T) {
	t.Parallel()

	parent, cancelParent := context.WithCancel(context.Background())
	cancelParent()
	startedAt := time.Now()
	stateCtx, cancelState := processorStateContext(parent)
	defer cancelState()

	require.NoError(t, stateCtx.Err())
	deadline, ok := stateCtx.Deadline()
	require.True(t, ok)
	require.WithinDuration(t, startedAt.Add(processorStateWriteTimeout), deadline, time.Second)
}
