package migrations

import (
	"strings"
	"testing"
)

const slackAssistantChannelConfigurationMigration = "000128_slack_assistant_channel_configuration"

func TestSlackAssistantChannelConfigurationMigrationContracts(t *testing.T) {
	t.Parallel()

	data, err := FS.ReadFile(slackAssistantChannelConfigurationMigration + ".up.sql")
	if err != nil {
		t.Fatalf("read forward migration: %v", err)
	}
	migration := string(data)

	for _, contract := range []string{
		"ADD COLUMN is_assistant_configured boolean NOT NULL DEFAULT false",
		"FROM public.slack_channel_team_access access",
		"SET is_assistant_configured = true",
		"AND team.workspace_id = access.workspace_id",
		"AND team.is_private = false",
		"AND access.slack_workspace_id = channel_record.slack_workspace_id",
		"AND access.slack_channel_id = channel_record.slack_channel_id",
	} {
		if !strings.Contains(migration, contract) {
			t.Errorf("migration is missing contract %q", contract)
		}
	}
	if strings.Contains(migration, "slack_channel_assistant_team_access") {
		t.Fatal("migration must keep the unified channel access table authoritative during rolling deploys")
	}
	for _, mutation := range []string{
		"DELETE FROM public.slack_channel_team_access",
		"INSERT INTO public.slack_channel_team_access",
		"UPDATE public.slack_channel_team_access",
	} {
		if strings.Contains(migration, mutation) {
			t.Errorf("migration must not mutate legacy access with %q", mutation)
		}
	}
}

func TestSlackAssistantChannelConfigurationRollbackPreservesUnifiedAccess(t *testing.T) {
	t.Parallel()

	data, err := FS.ReadFile(slackAssistantChannelConfigurationMigration + ".down.sql")
	if err != nil {
		t.Fatalf("read rollback migration: %v", err)
	}
	migration := string(data)

	if !strings.Contains(migration, "DROP COLUMN IF EXISTS is_assistant_configured") {
		t.Fatal("rollback must remove the assistant configuration marker")
	}
	if strings.Contains(migration, "DROP TABLE IF EXISTS public.slack_channel_team_access") {
		t.Fatal("rollback must preserve unified channel team access mappings")
	}
	if strings.Contains(migration, "slack_channel_assistant_team_access") {
		t.Fatal("rollback must not reference a staged assistant access table")
	}
}
