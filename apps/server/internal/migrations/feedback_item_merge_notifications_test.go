package migrations

import (
	"strings"
	"testing"
)

const feedbackItemMergeNotificationsMigration = "000125_feedback_item_merge_notifications"

func TestFeedbackItemMergeNotificationsForwardMigrationContract(t *testing.T) {
	t.Parallel()

	data, err := FS.ReadFile(feedbackItemMergeNotificationsMigration + ".up.sql")
	if err != nil {
		t.Fatalf("read forward migration: %v", err)
	}
	if !strings.Contains(string(data), "ADD VALUE IF NOT EXISTS 'feedback_item_merged'") {
		t.Fatal("forward migration must add the feedback_item_merged notification type")
	}
}

func TestFeedbackItemMergeNotificationsRollbackIsNonDestructive(t *testing.T) {
	t.Parallel()

	data, err := FS.ReadFile(feedbackItemMergeNotificationsMigration + ".down.sql")
	if err != nil {
		t.Fatalf("read rollback migration: %v", err)
	}
	if strings.Contains(strings.ToUpper(string(data)), "DROP TYPE") {
		t.Fatal("rollback must not rebuild the shared notification enum")
	}
}
