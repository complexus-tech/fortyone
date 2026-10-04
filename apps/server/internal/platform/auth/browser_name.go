package auth

import (
	"regexp"
	"strings"
)

const maxBrowserMetadataLength = 4096

var clientHintBrandPattern = regexp.MustCompile(`^"([^"\\]{1,64})"\s*;\s*v="[0-9.]{1,32}"$`)

// BrowserName reports a recognized browser brand from request metadata. It is
// display metadata, never authentication evidence. Chromium browsers that omit
// their own brand cannot be distinguished from the browser they report.
// Raw user agents, versions and device information are deliberately not stored.
func BrowserName(userAgent, clientHints string) *string {
	// Embedded app engines are not evidence that the user opened a browser,
	// even when their compatibility client hints include Chromium or Chrome.
	if len(userAgent) > maxBrowserMetadataLength || strings.Contains(userAgent, "Electron/") || strings.Contains(userAgent, "; wv)") {
		return nil
	}
	if len(clientHints) <= maxBrowserMetadataLength {
		brands := make(map[string]bool)
		for _, item := range strings.Split(clientHints, ",") {
			match := clientHintBrandPattern.FindStringSubmatch(strings.TrimSpace(item))
			if match != nil {
				brands[strings.ToLower(match[1])] = true
			}
		}
		// Specific brands precede their shared Chromium/Chrome compatibility brand.
		for _, brand := range []struct{ hint, name string }{
			{"dia", "Dia"}, {"arc", "Arc"}, {"opera", "Opera"},
			{"brave", "Brave"}, {"microsoft edge", "Edge"}, {"vivaldi", "Vivaldi"},
			{"google chrome", "Chrome"}, {"firefox", "Firefox"}, {"safari", "Safari"},
			{"chromium", "Chromium"},
		} {
			if brands[brand.hint] {
				return &brand.name
			}
		}
	}
	for _, brand := range []struct {
		tokens []string
		name   string
	}{
		{[]string{"Dia/"}, "Dia"}, {[]string{"Arc/"}, "Arc"},
		{[]string{"OPR/", "Opera/", "OPiOS/"}, "Opera"},
		{[]string{"Brave/"}, "Brave"}, {[]string{"Edg/", "EdgA/", "EdgiOS/", "Edge/"}, "Edge"},
		{[]string{"Vivaldi/"}, "Vivaldi"}, {[]string{"Firefox/", "FxiOS/"}, "Firefox"},
		{[]string{"Chromium/"}, "Chromium"}, {[]string{"Chrome/", "CriOS/"}, "Chrome"},
	} {
		for _, token := range brand.tokens {
			if hasBrowserProduct(userAgent, token) {
				return &brand.name
			}
		}
	}
	if hasBrowserProduct(userAgent, "Safari/") && hasBrowserProduct(userAgent, "Version/") {
		name := "Safari"
		return &name
	}
	return nil
}

func hasBrowserProduct(userAgent, token string) bool {
	for offset := 0; offset < len(userAgent); {
		found := strings.Index(userAgent[offset:], token)
		if found < 0 {
			return false
		}
		index := offset + found
		after := index + len(token)
		if (index == 0 || userAgent[index-1] == ' ' || userAgent[index-1] == '(') &&
			after < len(userAgent) && userAgent[after] >= '0' && userAgent[after] <= '9' {
			return true
		}
		offset = after
	}
	return false
}
