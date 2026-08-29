package migrations

import (
	"strings"
	"testing"
)

const messagingStoryMutationBatchProposalsMigration = "000129_messaging_story_mutation_batch_proposals"

func TestMessagingStoryMutationBatchProposalLifecycleContract(t *testing.T) {
	t.Parallel()

	data, err := FS.ReadFile(messagingStoryMutationBatchProposalsMigration + ".up.sql")
	if err != nil {
		t.Fatalf("read forward migration: %v", err)
	}
	migration := string(data)
	for _, contract := range []string{
		"operation = 'create_stories'",
		"status IN ('pending', 'applied')",
		"proposal IS NOT NULL",
		"status = 'applied'",
		"proposal IS NULL",
		"result IS NOT NULL",
		"last_error IS NULL",
		"status IN ('cancelled', 'expired')",
	} {
		if !strings.Contains(migration, contract) {
			t.Errorf("migration is missing proposal lifecycle contract %q", contract)
		}
	}
}
