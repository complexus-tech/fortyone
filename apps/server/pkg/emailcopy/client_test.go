package emailcopy

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (fn roundTripFunc) Do(request *http.Request) (*http.Response, error) {
	return fn(request)
}

func TestGenerateUsesLunaLowAndReturnsValidatedCopy(t *testing.T) {
	request := Request{
		Purpose:      "weekly strategy check-in",
		ProductVoice: "help the objective lead choose the next useful move",
		Facts: []Fact{
			{
				ReferenceID: "summary",
				Text:        "3 objectives are at risk and 8 objectives have not been updated in 7 days.",
				Required:    true,
			},
		},
		Actions: []Action{{ReferenceID: "strategy", Description: "Review strategy"}},
	}

	client, err := New(Config{
		APIKey: "test-key",
		HTTPClient: roundTripFunc(func(httpRequest *http.Request) (*http.Response, error) {
			require.Equal(t, "/v1/responses", httpRequest.URL.Path)
			require.Equal(t, "Bearer test-key", httpRequest.Header.Get("Authorization"))
			body, readErr := io.ReadAll(httpRequest.Body)
			require.NoError(t, readErr)
			var payload map[string]any
			require.NoError(t, json.Unmarshal(body, &payload))
			require.Equal(t, "gpt-5.6-luna", payload["model"])
			require.Equal(t, false, payload["store"])
			reasoning := payload["reasoning"].(map[string]any)
			require.Equal(t, "low", reasoning["effort"])
			require.Equal(t, "current_turn", reasoning["context"])
			format := payload["text"].(map[string]any)["format"].(map[string]any)
			require.Equal(t, true, format["strict"])

			return outputResponse(`{
				"subject":{"text":"3 objectives need a reset","referenceIds":["summary"]},
				"h1":{"text":"Bring 3 objectives back into focus","referenceIds":["summary"]},
				"intro":{"text":"A short review can reconnect today’s work with the outcomes that matter.","referenceIds":["summary"]},
				"senderProse":null,
				"rows":[{"referenceId":"summary","text":"3 objectives are at risk, and 8 have gone 7 days without an update.","ctaReferenceId":"strategy"}],
				"ctas":[{"referenceId":"strategy","label":"Review strategy"}],
				"feedbackThemeSummary":null,
				"replyPrompt":null
			}`), nil
		}),
	})
	require.NoError(t, err)

	output, err := client.Generate(context.Background(), request)

	require.NoError(t, err)
	require.Equal(t, "3 objectives need a reset", output.Subject.Text)
	require.Equal(t, "Review strategy", output.CTAs[0].Label)
}

func TestGenerateRejectsInventedNumericFacts(t *testing.T) {
	client, err := New(Config{
		APIKey: "test-key",
		HTTPClient: roundTripFunc(func(*http.Request) (*http.Response, error) {
			return outputResponse(`{
				"subject":{"text":"4 objectives need a reset","referenceIds":["summary"]},
				"h1":{"text":"Review 3 objectives","referenceIds":["summary"]},
				"intro":{"text":"A focused review will help.","referenceIds":["summary"]},
				"senderProse":null,
				"rows":[{"referenceId":"summary","text":"3 objectives are at risk.","ctaReferenceId":""}],
				"ctas":[],
				"feedbackThemeSummary":null,
				"replyPrompt":null
			}`), nil
		}),
	})
	require.NoError(t, err)

	_, err = client.Generate(context.Background(), Request{
		Purpose: "strategy check-in",
		Facts: []Fact{{
			ReferenceID: "summary",
			Text:        "3 objectives are at risk.",
			Required:    true,
		}},
	})

	require.ErrorContains(t, err, `invented numeric/date token "4"`)
}

func TestGeneratePreservesSignedNumericUnits(t *testing.T) {
	client, err := New(Config{
		APIKey: "test-key",
		HTTPClient: roundTripFunc(func(*http.Request) (*http.Response, error) {
			return outputResponse(`{
				"subject":{"text":"Progress needs attention","referenceIds":["progress"]},
				"h1":{"text":"Review current progress","referenceIds":["progress"]},
				"intro":{"text":"The current result needs a closer look.","referenceIds":["progress"]},
				"senderProse":null,
				"rows":[{"referenceId":"progress","text":"Activation is at 15% against a -10 target.","ctaReferenceId":""}],
				"ctas":[],
				"feedbackThemeSummary":null,
				"replyPrompt":null
			}`), nil
		}),
	})
	require.NoError(t, err)

	_, err = client.Generate(context.Background(), Request{
		Purpose: "key-result guidance",
		Facts: []Fact{{
			ReferenceID:  "progress",
			Text:         "Activation is at 15% against a -10% target.",
			EntityTokens: []string{"Activation"},
			Required:     true,
		}},
	})

	require.ErrorContains(t, err, `omitted protected numeric/date token "-10%"`)
}

func TestGenerateRejectsMissingProtectedEntity(t *testing.T) {
	client, err := New(Config{
		APIKey: "test-key",
		HTTPClient: roundTripFunc(func(*http.Request) (*http.Response, error) {
			return outputResponse(`{
				"subject":{"text":"An objective needs attention","referenceIds":["objective"]},
				"h1":{"text":"Reconnect the objective to this week’s work","referenceIds":["objective"]},
				"intro":{"text":"A quick review can clarify the next move.","referenceIds":["objective"]},
				"senderProse":null,
				"rows":[{"referenceId":"objective","text":"The revenue objective is at risk.","ctaReferenceId":""}],
				"ctas":[],
				"feedbackThemeSummary":null,
				"replyPrompt":null
			}`), nil
		}),
	})
	require.NoError(t, err)

	_, err = client.Generate(context.Background(), Request{
		Purpose: "strategy check-in",
		Facts: []Fact{{
			ReferenceID:  "objective",
			Text:         "The objective Grow enterprise revenue is at risk.",
			EntityTokens: []string{"Grow enterprise revenue"},
			Required:     true,
		}},
	})

	require.ErrorContains(t, err, `omitted protected token "Grow enterprise revenue"`)
}

func TestGenerateRejectsChangedProtectedCategoricalFact(t *testing.T) {
	client, err := New(Config{
		APIKey: "test-key",
		HTTPClient: roundTripFunc(func(*http.Request) (*http.Response, error) {
			return outputResponse(`{
				"subject":{"text":"Objective health needs attention","referenceIds":["objective"]},
				"h1":{"text":"Review the objective","referenceIds":["objective"]},
				"intro":{"text":"A review will clarify the next move.","referenceIds":["objective"]},
				"senderProse":null,
				"rows":[{"referenceId":"objective","text":"Improve onboarding is Off Track.","ctaReferenceId":""}],
				"ctas":[],
				"feedbackThemeSummary":null,
				"replyPrompt":null
			}`), nil
		}),
	})
	require.NoError(t, err)

	_, err = client.Generate(context.Background(), Request{
		Purpose: "strategy check-in",
		Facts: []Fact{{
			ReferenceID:     "objective",
			Text:            "Improve onboarding is At Risk.",
			EntityTokens:    []string{"Improve onboarding"},
			ProtectedTokens: []string{"At Risk"},
			Required:        true,
		}},
	})

	require.ErrorContains(t, err, `omitted protected token "At Risk"`)
}

func TestGenerateRejectsEntityFromUncitedFact(t *testing.T) {
	client, err := New(Config{
		APIKey: "test-key",
		HTTPClient: roundTripFunc(func(*http.Request) (*http.Response, error) {
			return outputResponse(`{
				"subject":{"text":"Make onboarding effortless needs attention","referenceIds":["task_a"]},
				"h1":{"text":"Review the work","referenceIds":["task_a"]},
				"intro":{"text":"One task needs a decision.","referenceIds":["task_a"]},
				"senderProse":null,
				"rows":[
					{"referenceId":"task_a","text":"Make onboarding effortless is overdue.","ctaReferenceId":""},
					{"referenceId":"task_b","text":"Ship audit logs is due tomorrow.","ctaReferenceId":""}
				],
				"ctas":[],
				"feedbackThemeSummary":null,
				"replyPrompt":null
			}`), nil
		}),
	})
	require.NoError(t, err)

	_, err = client.Generate(context.Background(), Request{
		Purpose: "task guidance",
		Facts: []Fact{
			{ReferenceID: "task_a", Text: "Ship audit logs is overdue.", EntityTokens: []string{"Ship audit logs"}, Required: true},
			{ReferenceID: "task_b", Text: "Make onboarding effortless is due tomorrow.", EntityTokens: []string{"Make onboarding effortless"}, Required: true},
		},
	})

	require.ErrorContains(t, err, `used protected token "Make onboarding effortless" without supporting fact`)
}

func TestValidateOutputRejectsSwappedCurrentAndTargetValues(t *testing.T) {
	request := Request{
		Purpose: "key-result guidance",
		Facts: []Fact{{
			ReferenceID:     "activation",
			Text:            "Activation; current value is 15%; target value is 40%.",
			EntityTokens:    []string{"Activation"},
			ProtectedTokens: []string{"current value is 15%", "target value is 40%"},
			Required:        true,
		}},
	}
	output := Output{
		Subject: GroundedText{
			Text:         "Activation: current value is 40%; target value is 15%",
			ReferenceIDs: []string{"activation"},
		},
		H1:    GroundedText{Text: "Give Activation a credible next move", ReferenceIDs: []string{"activation"}},
		Intro: GroundedText{Text: "A focused review can clarify what to change next.", ReferenceIDs: []string{"activation"}},
		Rows: []Row{{
			ReferenceID: "activation",
			Text:        "Activation; current value is 15%; target value is 40%.",
		}},
	}

	err := validateOutput(request, output)

	require.ErrorContains(t, err, "used numeric/date token")
	require.ErrorContains(t, err, "without one of its protected factual phrases")
}

func TestValidateOutputRejectsSwappedLabeledDates(t *testing.T) {
	request := Request{
		Purpose: "objective guidance",
		Facts: []Fact{{
			ReferenceID:     "objective",
			Text:            "Improve onboarding; last updated on August 1, 2026; ends on August 31, 2026.",
			EntityTokens:    []string{"Improve onboarding"},
			ProtectedTokens: []string{"last updated on August 1, 2026", "ends on August 31, 2026"},
			Required:        true,
		}},
	}
	output := Output{
		Subject: GroundedText{Text: "Improve onboarding needs a timely review", ReferenceIDs: []string{"objective"}},
		H1:      GroundedText{Text: "Reconnect the objective to the next milestone", ReferenceIDs: []string{"objective"}},
		Intro: GroundedText{
			Text:         "Improve onboarding was last updated on August 31, 2026 and ends on August 1, 2026.",
			ReferenceIDs: []string{"objective"},
		},
		Rows: []Row{{
			ReferenceID: "objective",
			Text:        "Improve onboarding; last updated on August 1, 2026; ends on August 31, 2026.",
		}},
	}

	err := validateOutput(request, output)

	require.ErrorContains(t, err, "used numeric/date token")
	require.ErrorContains(t, err, "without one of its protected factual phrases")
}

func TestValidateOutputAllowsCreativeProseAroundProtectedFacts(t *testing.T) {
	request := Request{
		Purpose: "key-result guidance",
		Facts: []Fact{{
			ReferenceID:     "activation",
			Text:            "Activation; current value is 15%; target value is 40%.",
			EntityTokens:    []string{"Activation"},
			ProtectedTokens: []string{"current value is 15%", "target value is 40%"},
			Required:        true,
		}},
	}
	output := Output{
		Subject: GroundedText{Text: "Activation needs a clearer next move", ReferenceIDs: []string{"activation"}},
		H1:      GroundedText{Text: "Turn the gap into a useful decision", ReferenceIDs: []string{"activation"}},
		Intro:   GroundedText{Text: "A focused review can help the team choose what to change next.", ReferenceIDs: []string{"activation"}},
		Rows: []Row{{
			ReferenceID: "activation",
			Text:        "Activation check: current value is 15%; target value is 40%. Decide which lever to test next.",
		}},
	}

	require.NoError(t, validateOutput(request, output))
}

func TestValidateOutputRequiresExplicitAIEmailReplyPrompt(t *testing.T) {
	request := Request{
		Purpose:            "task guidance",
		IncludeReplyPrompt: true,
		Facts: []Fact{{
			ReferenceID: "task",
			Text:        "One task needs attention.",
		}},
	}
	output := Output{
		Subject: GroundedText{Text: "One task needs attention", ReferenceIDs: []string{"task"}},
		H1:      GroundedText{Text: "Choose the next step", ReferenceIDs: []string{"task"}},
		Intro:   GroundedText{Text: "There is a decision to make.", ReferenceIDs: []string{"task"}},
		ReplyPrompt: &GroundedText{
			Text:         "I’m Maya, your AI agent. Reply to this email with the update you want.",
			ReferenceIDs: []string{"task"},
		},
	}

	require.NoError(t, validateOutput(request, output))

	output.ReplyPrompt.Text = "Tell me what you want updated."
	err := validateOutput(request, output)
	require.ErrorContains(t, err, `reply prompt must include "maya"`)
}

func TestValidateOutputRejectsSwappedProtectedActorAndStatusRoles(t *testing.T) {
	request := Request{
		Purpose: "feedback guidance",
		Facts: []Fact{{
			ReferenceID:     "feedback",
			Text:            "Feedback titled Export history was submitted by Amara and currently has status planned.",
			EntityTokens:    []string{"Export history"},
			ProtectedTokens: []string{"submitted by Amara", "currently has status planned"},
			Required:        true,
		}},
	}
	output := Output{
		Subject: GroundedText{Text: "Export history is ready for review", ReferenceIDs: []string{"feedback"}},
		H1:      GroundedText{Text: "Turn the feedback into a clear decision", ReferenceIDs: []string{"feedback"}},
		Intro:   GroundedText{Text: "This customer signal has a useful next step.", ReferenceIDs: []string{"feedback"}},
		Rows: []Row{{
			ReferenceID: "feedback",
			Text:        "Export history was submitted by planned and currently has status Amara.",
		}},
	}

	err := validateOutput(request, output)

	require.ErrorContains(t, err, `omitted protected token "submitted by Amara"`)
}

func TestGenerateIsDisabledWithoutAPIKey(t *testing.T) {
	client, err := New(Config{})
	require.NoError(t, err)

	_, err = client.Generate(context.Background(), Request{Purpose: "test", Facts: []Fact{{ReferenceID: "fact", Text: "A fact."}}})

	require.ErrorIs(t, err, ErrNotConfigured)
}

func TestGenerateRejectsProtectedEntityTokenMissingFromSourceFact(t *testing.T) {
	client, err := New(Config{APIKey: "test-key"})
	require.NoError(t, err)

	_, err = client.Generate(context.Background(), Request{
		Purpose: "strategy check-in",
		Facts: []Fact{{
			ReferenceID:  "objective",
			Text:         "An objective is at risk.",
			EntityTokens: []string{"Improve onboarding"},
			Required:     true,
		}},
	})

	require.ErrorContains(t, err, `does not contain protected token "Improve onboarding"`)
}

func outputResponse(output string) *http.Response {
	encoded, _ := json.Marshal(map[string]any{
		"output": []any{map[string]any{
			"content": []any{map[string]any{
				"type": "output_text",
				"text": output,
			}},
		}},
	})
	return &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(strings.NewReader(string(encoded))),
		Header:     make(http.Header),
	}
}
