package slack

import "strings"

var slackRequiredBotOAuthScopes = [...]string{
	"app_mentions:read",
	"channels:history",
	"channels:read",
	"chat:write",
	"chat:write.public",
	"commands",
	"groups:history",
	"groups:read",
	"im:history",
	"links:read",
	"links:write",
}

// slackBotOAuthScopeValue is the canonical scope value shared by the runtime
// OAuth flow and the source-controlled Slack manifest.
func slackBotOAuthScopeValue() string {
	return strings.Join(slackRequiredBotOAuthScopes[:], ",")
}
