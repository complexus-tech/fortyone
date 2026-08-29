package usershttp

import "testing"

func TestSanitizeCallbackURL(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name             string
		raw              string
		configuredDomain string
		websiteURL       string
		want             string
		wantError        bool
	}{
		{
			name:       "relative application path",
			raw:        "/auth-callback?callbackUrl=%2Ffeedback",
			websiteURL: "https://cloud.fortyone.app",
			want:       "/auth-callback?callbackUrl=%2Ffeedback",
		},
		{
			name:             "workspace subdomain",
			raw:              "https://art-circles.fortyone.app/feedback",
			configuredDomain: ".fortyone.app",
			websiteURL:       "https://cloud.fortyone.app",
			want:             "https://art-circles.fortyone.app/feedback",
		},
		{
			name:       "local development host",
			raw:        "http://art-circles.localhost:3000/feedback",
			websiteURL: "http://localhost:3000",
			want:       "http://art-circles.localhost:3000/feedback",
		},
		{
			name:       "protocol relative URL",
			raw:        "//malicious.example.com",
			websiteURL: "https://cloud.fortyone.app",
			wantError:  true,
		},
		{
			name:       "backslash relative URL",
			raw:        `/\\malicious.example.com`,
			websiteURL: "https://cloud.fortyone.app",
			wantError:  true,
		},
		{
			name:       "hostile suffix",
			raw:        "https://fortyone.app.malicious.example.com/feedback",
			websiteURL: "https://cloud.fortyone.app",
			wantError:  true,
		},
		{
			name:       "local callback in production",
			raw:        "http://localhost:3000/feedback",
			websiteURL: "https://cloud.fortyone.app",
			wantError:  true,
		},
		{
			name:       "embedded credentials",
			raw:        "https://fortyone.app@malicious.example.com/feedback",
			websiteURL: "https://cloud.fortyone.app",
			wantError:  true,
		},
		{
			name:       "insecure production URL",
			raw:        "http://cloud.fortyone.app/auth-callback",
			websiteURL: "https://cloud.fortyone.app",
			wantError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := sanitizeCallbackURL(tt.raw, tt.configuredDomain, tt.websiteURL)
			if tt.wantError {
				if err == nil {
					t.Fatalf("sanitizeCallbackURL() error = nil, want error")
				}
				return
			}
			if err != nil {
				t.Fatalf("sanitizeCallbackURL() error = %v", err)
			}
			if got != tt.want {
				t.Fatalf("sanitizeCallbackURL() = %q, want %q", got, tt.want)
			}
		})
	}
}
