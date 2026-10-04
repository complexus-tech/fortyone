package auth

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestBrowserNameRecognizesReportedBrands(t *testing.T) {
	t.Parallel()
	chrome := "Mozilla/5.0 AppleWebKit/537.36 Chrome/140.0.0.0 Safari/537.36"
	for _, test := range []struct{ name, agent, hints, want string }{
		{"Chrome user agent", chrome, "", "Chrome"},
		{"Chrome hints", chrome, `"Not A;Brand";v="99", "Chromium";v="140", "Google Chrome";v="140"`, "Chrome"},
		{"Opera compatibility tokens", chrome + " OPR/123.0.0.0", "", "Opera"},
		{"Opera brand", chrome, `"Opera";v="123", "Chromium";v="140"`, "Opera"},
		{"Arc reported brand", chrome, `"Chromium";v="140", "Google Chrome";v="140", "Arc";v="1"`, "Arc"},
		{"Dia reported brand", chrome, `"Dia";v="1", "Chromium";v="140"`, "Dia"},
		{"Dia explicit product", chrome + " Dia/1.0", "", "Dia"},
		{"Arc explicit product", chrome + " Arc/1.0", "", "Arc"},
		{"Edge", chrome + " Edg/140.0", "", "Edge"},
		{"Firefox", "Mozilla/5.0 Gecko/20100101 Firefox/143.0", "", "Firefox"},
		{"Safari", "Mozilla/5.0 AppleWebKit/605.1.15 Version/18.6 Safari/605.1.15", "", "Safari"},
		{"Chrome iOS", "Mozilla/5.0 AppleWebKit/605.1.15 CriOS/140.0 Mobile/15E148 Safari/604.1", "", "Chrome"},
		{"Chromium engine only", chrome, `"Chromium";v="140", "Not A;Brand";v="99"`, "Chromium"},
		{"case insensitive hints", chrome, `"opera";v="123"`, "Opera"},
		{"malformed hints use agent", chrome, `"Opera"`, "Chrome"},
		{"unknown", "FortyOne/1.0", "", ""},
		{"empty historical metadata", "", "", ""},
		{"engine Safari token only", "AppleWebKit/605.1.15 Safari/605.1.15", "", ""},
		{"embedded native engine", chrome + " Electron/33.0", "", ""},
		{"embedded engine hints", chrome + " Electron/33.0", `"Chromium";v="140", "Google Chrome";v="140"`, ""},
		{"Android webview", "Mozilla/5.0 (Linux; Android 15; wv) AppleWebKit/537.36 Version/4.0 Chrome/140.0.0.0 Mobile Safari/537.36", "", ""},
		{"Android webview hints", "Mozilla/5.0 (Linux; Android 15; wv) Chrome/140.0.0.0", `"Android WebView";v="140", "Chromium";v="140"`, ""},
		{"lookalike product", "NotChrome/140.0 NotArc/1.0", "", ""},
		{"nested malformed product", "NotChrome/Chrome/140.0", "", ""},
		{"arbitrary hint excluded", "", `"My private device name";v="140"`, ""},
		{"bounded agent", strings.Repeat("x", maxBrowserMetadataLength) + " Chrome/140.0", "", ""},
		{"bounded hints use agent", chrome, strings.Repeat("x", maxBrowserMetadataLength) + `,"Arc";v="1"`, "Chrome"},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			got := BrowserName(test.agent, test.hints)
			if test.want == "" {
				require.Nil(t, got)
			} else {
				require.NotNil(t, got)
				require.Equal(t, test.want, *got)
			}
		})
	}
}
