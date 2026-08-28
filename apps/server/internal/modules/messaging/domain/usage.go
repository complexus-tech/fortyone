package messagingdomain

// Usage contains aggregate token usage for quota accounting.
type Usage struct {
	InputTokens  int
	OutputTokens int
	TotalTokens  int
}
