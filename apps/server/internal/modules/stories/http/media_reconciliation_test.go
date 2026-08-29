package storieshttp

import (
	"encoding/json"
	"errors"
	"testing"

	stories "github.com/complexus-tech/projects-api/internal/modules/stories/service"
	"github.com/google/uuid"
)

func TestStoryMediaReconciliationRequestDefaultsToDisabled(t *testing.T) {
	enabled, references, err := storyMediaReconciliationRequest(
		map[string]json.RawMessage{"descriptionHTML": json.RawMessage(`"<p>Draft</p>"`)},
		"acme",
		uuid.New(),
	)
	if err != nil || enabled || references != nil {
		t.Fatalf("enabled=%v references=%v err=%v", enabled, references, err)
	}
}

func TestStoryMediaReconciliationRequestRequiresDescriptionHTML(t *testing.T) {
	_, _, err := storyMediaReconciliationRequest(
		map[string]json.RawMessage{reconcileDescriptionMediaField: json.RawMessage("true")},
		"acme",
		uuid.New(),
	)
	if err == nil {
		t.Fatal("expected an authoritative media snapshot without HTML to fail")
	}
}

func TestParseStoryMediaReferencesAcceptsOnlyStableMatchingMedia(t *testing.T) {
	storyID := uuid.New()
	imageID := uuid.New()
	videoID := uuid.New()
	contentHTML := `<p>Before</p>` +
		`<img src="/workspaces/acme/stories/` + storyID.String() + `/media/` + imageID.String() + `" data-attachment-id="` + imageID.String() + `">` +
		`<img src="https://api.example.test/workspaces/acme/stories/` + storyID.String() + `/media/` + imageID.String() + `" data-attachment-id="` + imageID.String() + `">` +
		`<video data-document-media-video="true" src="https://api.example.test/workspaces/acme/stories/` + storyID.String() + `/media/` + videoID.String() + `" data-attachment-id="` + videoID.String() + `"></video>` +
		`<img src="https://images.example.test/external.png">`

	references, err := parseStoryMediaReferences(contentHTML, "acme", storyID)
	if err != nil {
		t.Fatalf("parse media references: %v", err)
	}
	if len(references) != 2 || references[0] != imageID || references[1] != videoID {
		t.Fatalf("references=%v, want [%s %s]", references, imageID, videoID)
	}
}

func TestParseStoryMediaReferencesRejectsMismatchedStableMedia(t *testing.T) {
	storyID := uuid.New()
	attachmentID := uuid.New()
	tests := map[string]string{
		"attachment id":        `<img src="/workspaces/acme/stories/` + storyID.String() + `/media/` + attachmentID.String() + `" data-attachment-id="` + uuid.NewString() + `">`,
		"invalid absolute url": `<img src="https:/workspaces/acme/stories/` + storyID.String() + `/media/` + attachmentID.String() + `" data-attachment-id="` + attachmentID.String() + `">`,
		"query":                `<img src="/workspaces/acme/stories/` + storyID.String() + `/media/` + attachmentID.String() + `?token=unsafe" data-attachment-id="` + attachmentID.String() + `">`,
		"story":                `<img src="/workspaces/acme/stories/` + uuid.NewString() + `/media/` + attachmentID.String() + `" data-attachment-id="` + attachmentID.String() + `">`,
		"workspace":            `<img src="/workspaces/other/stories/` + storyID.String() + `/media/` + attachmentID.String() + `" data-attachment-id="` + attachmentID.String() + `">`,
	}

	for name, contentHTML := range tests {
		t.Run(name, func(t *testing.T) {
			_, err := parseStoryMediaReferences(contentHTML, "acme", storyID)
			if !errors.Is(err, stories.ErrInvalidStoryMediaReference) {
				t.Fatalf("error=%v, want invalid story media reference", err)
			}
		})
	}
}
