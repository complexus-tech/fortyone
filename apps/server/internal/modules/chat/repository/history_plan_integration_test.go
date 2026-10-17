//go:build integration

package chatrepository

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	domain "github.com/complexus-tech/projects-api/internal/modules/chat/domain"
	chatsql "github.com/complexus-tech/projects-api/internal/modules/chat/repository/sqlc"
	chatservice "github.com/complexus-tech/projects-api/internal/modules/chat/service"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/require"
)

type explainedQuery struct {
	sql  string
	plan []byte
}
type historyPlanRecorder struct {
	pgx.Tx
	plans []explainedQuery
}

func (r *historyPlanRecorder) Query(ctx context.Context, query string, args ...any) (pgx.Rows, error) {
	var plan []byte
	if err := r.Tx.QueryRow(ctx, "EXPLAIN (ANALYZE, FORMAT JSON) "+query, args...).Scan(&plan); err != nil {
		return nil, err
	}
	r.plans = append(r.plans, explainedQuery{sql: query, plan: plan})
	return r.Tx.Query(ctx, query, args...)
}

func TestChatHistoryUsesBoundedIndexesWithGenericPreparedPlans(t *testing.T) {
	f := newFixture(t)
	_, err := f.pool.Exec(f.ctx, `INSERT INTO chat_conversation_messages(workspace_id,conversation_id,sequence,author_id,creation_id,creation_hash,text)
        SELECT $1,$2,n,$3,gen_random_uuid(),decode(repeat('ab',32),'hex'),'Historical root ' || n FROM generate_series(1,10000) n`, f.workspace, f.room, f.author)
	require.NoError(t, err)
	var root uuid.UUID
	require.NoError(t, f.pool.QueryRow(f.ctx, `SELECT id FROM chat_conversation_messages WHERE conversation_id=$1 AND sequence=1`, f.room).Scan(&root))
	_, err = f.pool.Exec(f.ctx, `INSERT INTO chat_conversation_messages(workspace_id,conversation_id,sequence,thread_root_id,author_id,creation_id,creation_hash,text)
        SELECT $1,$2,10000+n,$3,$4,gen_random_uuid(),decode(repeat('cd',32),'hex'),'Historical reply ' || n FROM generate_series(1,100) n`, f.workspace, f.room, root, f.author)
	require.NoError(t, err)
	_, err = f.pool.Exec(f.ctx, `ANALYZE public.chat_conversation_messages`)
	require.NoError(t, err)
	tx, err := f.pool.Begin(f.ctx)
	require.NoError(t, err)
	defer func() { _ = tx.Rollback(f.ctx) }()
	_, err = tx.Exec(f.ctx, `SET LOCAL plan_cache_mode='force_generic_plan'`)
	require.NoError(t, err)
	recorder := &historyPlanRecorder{Tx: tx}
	repo := &Repository{pool: f.pool, queries: chatsql.New(recorder)}
	service := chatservice.New(repo)
	roots, err := service.ListMessages(f.ctx, domain.MessageQuery{Scope: f.scope, ConversationID: f.room, Limit: 50})
	require.NoError(t, err)
	require.Len(t, roots.Items, 50)
	require.EqualValues(t, 9951, roots.Items[0].Sequence)
	replies, err := service.ListMessages(f.ctx, domain.MessageQuery{Scope: f.scope, ConversationID: f.room, ThreadRootID: &root, Limit: 50})
	require.NoError(t, err)
	require.Len(t, replies.Items, 50)
	require.EqualValues(t, 10001, replies.Items[0].Sequence)
	require.Len(t, recorder.plans, 2)
	for _, entry := range recorder.plans {
		var plans []struct {
			Plan          map[string]any `json:"Plan"`
			ExecutionTime float64        `json:"Execution Time"`
		}
		require.NoError(t, json.Unmarshal(entry.plan, &plans))
		require.Len(t, plans, 1)
		indexed := false
		var inspect func(map[string]any)
		inspect = func(node map[string]any) {
			if name, ok := node["Index Name"].(string); ok && (strings.Contains(name, "chat_roots_sequence") || strings.Contains(name, "chat_replies_sequence") || strings.Contains(name, "conversation_id_sequence")) {
				indexed = true
			}
			if kind, ok := node["Node Type"].(string); ok && kind == "Sort" {
				require.LessOrEqual(t, node["Actual Rows"].(float64), float64(50), "history sorting must be limited to the returned window")
			}
			if children, ok := node["Plans"].([]any); ok {
				for _, child := range children {
					inspect(child.(map[string]any))
				}
			}
		}
		inspect(plans[0].Plan)
		require.True(t, indexed, "history query must traverse a conversation/thread sequence index")
		t.Logf("%s over 10,100 messages: %.2f ms", strings.SplitN(entry.sql, "\n", 2)[0], plans[0].ExecutionTime)
	}
}
