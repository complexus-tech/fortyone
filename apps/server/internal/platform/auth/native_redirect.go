package auth

import "github.com/complexus-tech/projects-api/internal/platform/deployment"

const (
	DesktopRedirectURI            = "fortyone-desktop://login"
	DesktopDevelopmentRedirectURI = "fortyone-desktop-dev://login"
)

// IsAllowedDesktopRedirect keeps native handoffs on exact, registered callback
// addresses. Development builds use a separate scheme so they do not activate
// an installed production app. No other deployment mode enables that scheme.
func IsAllowedDesktopRedirect(value string, mode deployment.Mode) bool {
	return value == DesktopRedirectURI || (mode == deployment.Development && value == DesktopDevelopmentRedirectURI)
}
