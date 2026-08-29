package emailcopy

import "context"

// Generator produces grounded, product-quality email copy from caller-owned
// facts and actions. Callers remain responsible for resolving reference IDs to
// trusted labels and URLs.
type Generator interface {
	Generate(ctx context.Context, request Request) (Output, error)
}

// Request describes one email without exposing executable actions or URLs to
// the model.
type Request struct {
	SafetyIdentifier            string   `json:"-"`
	Purpose                     string   `json:"purpose"`
	ProductVoice                string   `json:"productVoice"`
	Facts                       []Fact   `json:"facts"`
	Actions                     []Action `json:"actions"`
	IncludeSenderProse          bool     `json:"includeSenderProse"`
	IncludeFeedbackThemeSummary bool     `json:"includeFeedbackThemeSummary"`
	IncludeReplyPrompt          bool     `json:"includeReplyPrompt"`
}

// Fact is an immutable statement the generated copy may rely on. Required
// facts must receive a dedicated row and retain their protected literal and
// numeric/date tokens. Protected tokens should include entity names and any
// categorical values whose meaning must not be paraphrased, such as status,
// health, measurement units, or reporting periods.
type Fact struct {
	ReferenceID     string   `json:"referenceId"`
	Text            string   `json:"text"`
	EntityTokens    []string `json:"entityTokens"`
	ProtectedTokens []string `json:"protectedTokens"`
	Required        bool     `json:"required"`
}

// Action identifies a trusted destination owned by the caller. Description is
// context for choosing an appropriate label; the model never receives a URL.
type Action struct {
	ReferenceID string `json:"referenceId"`
	Description string `json:"description"`
}

type GroundedText struct {
	Text         string   `json:"text"`
	ReferenceIDs []string `json:"referenceIds"`
}

type Row struct {
	ReferenceID    string `json:"referenceId"`
	Text           string `json:"text"`
	CTAReferenceID string `json:"ctaReferenceId"`
}

type CTA struct {
	ReferenceID string `json:"referenceId"`
	Label       string `json:"label"`
}

// Output contains every visible prose surface used by a product email. The
// nullable sections are permitted only when their matching request flags are
// enabled.
type Output struct {
	Subject              GroundedText  `json:"subject"`
	H1                   GroundedText  `json:"h1"`
	Intro                GroundedText  `json:"intro"`
	SenderProse          *GroundedText `json:"senderProse"`
	Rows                 []Row         `json:"rows"`
	CTAs                 []CTA         `json:"ctas"`
	FeedbackThemeSummary *GroundedText `json:"feedbackThemeSummary"`
	ReplyPrompt          *GroundedText `json:"replyPrompt"`
}
