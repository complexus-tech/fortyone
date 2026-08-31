package chatsessionsrepository

import (
	"os"
	"strings"
	"testing"
)

func TestMessageCountSubtractsCurrentPeriodAIUsageReset(t *testing.T) {
	t.Parallel()

	raw, err := os.ReadFile("queries/sessions.sql")
	if err != nil {
		t.Fatal(err)
	}
	query := strings.Join(strings.Fields(string(raw)), " ")
	for _, contract := range []string{
		"FROM public.user_ai_usage_resets AS reset",
		"reset.period_start = sqlc.arg(start_date)",
		") - COALESCE((",
		"GREATEST(",
		"session.deleted_at IS NULL",
	} {
		if !strings.Contains(query, contract) {
			t.Errorf("message count query is missing reset contract %q", contract)
		}
	}
}
