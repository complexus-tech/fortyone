//go:build integration && chatload

package api

import (
	"fmt"
	"net/http"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// The admin settings endpoint and durable-claim/SSE gating have focused
// integration coverage. This post-load probe verifies that both already-running
// API replicas observe the same live database setting without process restart.
// The direct setting writes are restricted to the runner's disposable database.
func (r *loadRunner) validateWorkspaceFeature(pool *pgxpool.Pool) error {
	for _, enabled := range []bool{false, true} {
		_, err := pool.Exec(r.ctx, `
			INSERT INTO workspace_settings (workspace_id, chat_enabled)
			VALUES ($1, $2)
			ON CONFLICT (workspace_id) DO UPDATE SET chat_enabled=EXCLUDED.chat_enabled`, r.workspace, enabled)
		if err != nil {
			return err
		}
		for user := range r.servers {
			var capability struct {
				Enabled bool `json:"enabled"`
			}
			if err := r.request(user, http.MethodGet, r.path("capabilities"), "feature_capabilities", nil, &capability); err != nil {
				return err
			}
			if capability.Enabled != enabled {
				return fmt.Errorf("API replica %d did not observe chat_enabled=%t", user, enabled)
			}
			status, operation := http.StatusOK, "feature_restored_read"
			if !enabled {
				status, operation = http.StatusForbidden, "feature_disabled_read"
			}
			path := r.path("conversations/" + r.rooms[r.people[user].Team].String() + "/messages")
			if err := r.requestStatus(user, http.MethodGet, path, operation, nil, nil, status); err != nil {
				return err
			}
			if !enabled {
				input := map[string]any{"creationId": uuid.New(), "text": "Disabled workspace probe"}
				if err := r.requestStatus(user, http.MethodPost, path, "feature_disabled_create", input, nil, http.StatusForbidden); err != nil {
					return err
				}
			}
		}
	}
	return nil
}
