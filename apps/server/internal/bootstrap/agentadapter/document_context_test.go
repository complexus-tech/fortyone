package agentadapter

import (
	"testing"

	messaging "github.com/complexus-tech/projects-api/internal/modules/messaging/service"
	"github.com/complexus-tech/projects-api/internal/platform/documentmentions"
	"github.com/google/uuid"
)

func TestDocumentDirectCommandsDoNotTreatQuestionsOrQuotedInstructionsAsApproval(t *testing.T) {
	for _, test := range []struct {
		body string
		want bool
	}{
		{"[@Maya](agent:maya) Please rename this document to Plan", true},
		{"[@Maya](agent:maya) What should I change?", false},
		{"> Rename this document", false},
		{"`Rename this document`", false},
		{"Could you suggest a new title?", false},
	} {
		if got := directDocumentCommand(test.body); got != test.want {
			t.Errorf("%q got%v want%v", test.body, got, test.want)
		}
	}
}

func TestDocumentReviewCannotNotifyMentionsCopiedFromDocumentText(t *testing.T) {
	id := uuid.New()
	text := "```\n[@Copied](mention:" + id.String() + ")\n```"
	review := documentReview(messaging.DocumentOperations{Version: 1, Title: &text, Replacements: []messaging.DocumentReplacement{{Before: text, After: "replacement"}}, AppendParagraphs: []string{text}})
	if ids := documentmentions.Candidates(review).UserIDs; len(ids) != 0 {
		t.Fatalf("review could notify copied document text: %v", ids)
	}
}
