package storiesrepository

import (
	"testing"

	stories "github.com/complexus-tech/projects-api/internal/modules/stories/service"
	"github.com/google/uuid"
)

func TestShouldCompactActivityOnlyCompactsUpdates(t *testing.T) {
	if !shouldCompactActivity(stories.CoreActivity{Type: "update", Field: "status_id"}) {
		t.Fatal("expected update activity to be compactable")
	}

	if shouldCompactActivity(stories.CoreActivity{Type: "create", Field: "story"}) {
		t.Fatal("expected create activity not to be compactable")
	}

	if shouldCompactActivity(stories.CoreActivity{Type: "link", Field: "url"}) {
		t.Fatal("expected link activity not to be compactable")
	}

	if shouldCompactActivity(stories.CoreActivity{Type: "update"}) {
		t.Fatal("expected update activity without a field not to be compactable")
	}

	id := uuid.New()
	activity := stories.CoreActivity{ID: id, Type: "update", Field: "auto_scheduling_status"}
	if shouldCompactActivity(activity) {
		t.Fatal("expected a source-identified activity to bypass lossy compaction")
	}
	if converted := toDBActivity(activity); converted.ID != id {
		t.Fatalf("activity id was not preserved for idempotent insertion: got %s want %s", converted.ID, id)
	}
}
