package slack

import "strings"

var slackRequiredBotOAuthScopes = [...]string{
	"app_mentions:read",
	"channels:history",
	"channels:read",
	"chat:write",
	"chat:write.public",
	"commands",
	"files:read",
	"groups:history",
	"groups:read",
	"im:history",
	"links:read",
	"links:write",
	"mpim:history",
	"users:read",
	"users:read.email",
}

// slackBotOAuthScopeValue is the canonical scope value shared by the runtime
// OAuth flow and the source-controlled Slack manifest.
func slackBotOAuthScopeValue() string {
	return strings.Join(slackRequiredBotOAuthScopes[:], ",")
}

func slackBotHasScope(granted *string, required string) bool {
	if granted == nil {
		return false
	}
	for _, scope := range strings.Split(*granted, ",") {
		if strings.TrimSpace(scope) == required {
			return true
		}
	}
	return false
}
