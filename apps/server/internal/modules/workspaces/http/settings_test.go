package workspaceshttp

import (
	"encoding/json"
	"testing"
	"time"

	workspaces "github.com/complexus-tech/projects-api/internal/modules/workspaces/service"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestChatSettingPartialUpdatesPreserveOtherFeatures(t *testing.T) {
	workspace := uuid.New()
	for _, initial := range []bool{true, false} {
		current := workspaces.CoreWorkspaceSettings{
			UpdatedAt:   time.Date(2026, 10, 17, 14, 0, 0, 123000, time.UTC),
			ChatEnabled: initial, ObjectiveEnabled: true, KeyResultEnabled: false,
			WorkingDays: []int{1, 2, 3, 4, 5}, WorkingStartMinute: 540, WorkingEndMinute: 1020,
		}
		for _, test := range []struct {
			payload string
			want    bool
		}{
			{payload: `{}`, want: initial},
			{payload: `{"chatEnabled":true}`, want: true},
			{payload: `{"chatEnabled":false}`, want: false},
		} {
			var input AppUpdateWorkspaceSettings
			require.NoError(t, json.Unmarshal([]byte(test.payload), &input))
			updated := toCoreWorkspaceSettings(input, workspace, current)
			require.Equal(t, test.want, updated.ChatEnabled, test.payload)
			require.Equal(t, current.UpdatedAt, updated.UpdatedAt, "the DB read version travels with the merged partial update")
			require.True(t, updated.ObjectiveEnabled)
			require.False(t, updated.KeyResultEnabled)
			require.Equal(t, current.WorkingDays, updated.WorkingDays)
			require.Equal(t, current.WorkingStartMinute, updated.WorkingStartMinute)
			require.Equal(t, current.WorkingEndMinute, updated.WorkingEndMinute)
			response, err := json.Marshal(toAppWorkspaceSettings(updated))
			require.NoError(t, err)
			var fields map[string]any
			require.NoError(t, json.Unmarshal(response, &fields))
			require.Equal(t, test.want, fields["chatEnabled"])
		}
	}
}
