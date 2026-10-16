package agentinvocationshttp

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	agentruns "github.com/complexus-tech/projects-api/internal/modules/agentinvocations/service"
	"github.com/google/uuid"
)

func TestApprovalPresentationExposesOnlyReviewText(t *testing.T) {
	run := agentruns.Run{Status: agentruns.StatusAwaitingApproval,
		Proposal: &agentruns.Proposal{Review: "Description: Before → After", Action: json.RawMessage(`{"token":"signed-secret"}`)},
		Outcome:  &agentruns.Outcome{Receipt: json.RawMessage(`{"private":"receipt"}`)},
	}
	for _, status := range []agentruns.Status{agentruns.StatusAwaitingApproval, agentruns.StatusQueued, agentruns.StatusSucceeded} {
		run.Status = status
		body, err := json.Marshal(present(run))
		if err != nil {
			t.Fatal(err)
		}
		for _, secret := range []string{"signed-secret", "receipt", "action", "token"} {
			if strings.Contains(string(body), secret) {
				t.Fatalf("response exposes %q: %s", secret, body)
			}
		}
		if got := strings.Contains(string(body), "proposalReview"); got != (status == agentruns.StatusAwaitingApproval) {
			t.Fatalf("review exposure for status %s: %s", status, body)
		}
	}
}

func TestPresentationExposesCommittedReplyDuringDeliveryRetries(t *testing.T) {
	proposalID, failureID, resultID := uuid.New(), uuid.New(), uuid.New()
	approvedAt := time.Now()
	for _, status := range []agentruns.Status{agentruns.StatusQueued, agentruns.StatusRunning} {
		t.Run(string(status), func(t *testing.T) {
			for _, test := range []struct {
				name string
				run  agentruns.Run
				want *uuid.UUID
			}{
				{
					name: "approval question awaiting delivery",
					run: agentruns.Run{Proposal: &agentruns.Proposal{RequiresApproval: true},
						ProposalReplyID: &proposalID},
					want: &proposalID,
				},
				{
					name: "failure reply awaiting delivery",
					run:  agentruns.Run{FailureReplyID: &failureID, FailurePending: true},
					want: &failureID,
				},
				{
					name: "result awaiting delivery",
					run:  agentruns.Run{ReplyCommentID: &resultID},
					want: &resultID,
				},
				{
					name: "question not published yet",
					run:  agentruns.Run{Proposal: &agentruns.Proposal{RequiresApproval: true}},
				},
				{
					name: "failure not published yet",
					run:  agentruns.Run{FailurePending: true},
				},
				{
					name: "approved action still working",
					run: agentruns.Run{Proposal: &agentruns.Proposal{RequiresApproval: true},
						ProposalReplyID: &proposalID, ApprovedAt: &approvedAt},
				},
				{
					name: "retried failure still working",
					run:  agentruns.Run{FailureReplyID: &failureID},
				},
				{
					name: "new result takes precedence over earlier replies",
					run: agentruns.Run{ReplyCommentID: &resultID, ProposalReplyID: &proposalID,
						FailureReplyID: &failureID},
					want: &resultID,
				},
			} {
				t.Run(test.name, func(t *testing.T) {
					test.run.Status = status
					got := present(test.run).ReplyCommentID
					if (got == nil) != (test.want == nil) || (got != nil && *got != *test.want) {
						t.Fatalf("replyCommentId = %v, want %v", got, test.want)
					}
				})
			}
		})
	}
}
