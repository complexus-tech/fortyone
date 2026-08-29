package usersrepository

import (
	"os"
	"strings"
	"testing"
)

func TestBrowserSessionQueriesPreserveMonotonicRevocation(t *testing.T) {
	t.Parallel()

	browserSource := normalizedQuerySource(t, "queries/browser_sessions.sql")
	for _, clause := range []string{
		"select account.auth_session_version",
		"where account.user_id = sqlc.arg(user_id)",
		"and account.is_active = true",
	} {
		if !strings.Contains(browserSource, clause) {
			t.Fatalf("browser session query is missing %q", clause)
		}
	}

	accountSource := normalizedQuerySource(t, "queries/accounts.sql")
	deactivation := namedQueryContract(t, accountSource, "-- name: deactivateuser :execrows", "-- name: updatelastusedworkspaceformember :execrows")
	if !strings.Contains(deactivation, "auth_session_version = auth_session_version + 1") {
		t.Fatal("self-deactivation does not increment the browser session version")
	}
	if !strings.Contains(deactivation, "login_reactivation_policy = 'verified_sign_in'") {
		t.Fatal("self-deactivation does not preserve verified sign-in reactivation")
	}

	activation := namedQueryContract(t, accountSource, "-- name: reactivateuserforverifiedsignin :one", "-- name: deactivateuser :execrows")
	if strings.Contains(activation, "auth_session_version") {
		t.Fatal("reactivation must never reset or decrement the browser session version")
	}
	for _, clause := range []string{
		"and is_active = false",
		"and login_reactivation_policy = 'verified_sign_in'",
		"last_login_at = cast(sqlc.arg(signed_in_at) as timestamptz)",
		"updated_at = cast(sqlc.arg(signed_in_at) as timestamptz)",
	} {
		if !strings.Contains(activation, clause) {
			t.Fatalf("verified sign-in reactivation is missing %q", clause)
		}
	}
}

func normalizedQuerySource(t *testing.T, path string) string {
	t.Helper()
	source, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	return strings.Join(strings.Fields(strings.ToLower(string(source))), " ")
}

func namedQueryContract(t *testing.T, source, start, end string) string {
	t.Helper()
	startIndex := strings.Index(source, start)
	endIndex := strings.Index(source, end)
	if startIndex < 0 || endIndex <= startIndex {
		t.Fatalf("query boundaries %q..%q were not found", start, end)
	}
	return source[startIndex:endIndex]
}
