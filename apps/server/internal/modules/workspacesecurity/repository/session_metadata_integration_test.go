//go:build integration

package workspacesecurityrepository_test

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSessionRegistryPreservesIssuanceBrowserAndUnknownHistory(t *testing.T) {
	f := seed(t)
	member := f.scope
	member.ActorID = f.member
	known := identity()
	chrome := "Chrome"
	known.BrowserName = &chrome
	require.NoError(t, f.repo.CheckSession(t.Context(), member, known))
	safari := "Safari"
	known.BrowserName = &safari
	require.NoError(t, f.repo.CheckSession(t.Context(), member, known))
	unknown := identity()
	require.NoError(t, f.repo.CheckSession(t.Context(), member, unknown))
	require.NoError(t, f.repo.CheckSession(t.Context(), member, unknown))
	rows, err := f.repo.Sessions(t.Context(), f.scope, &f.member, false)
	require.NoError(t, err)
	require.Len(t, rows.Items, 2)
	for _, row := range rows.Items {
		require.Equal(t, f.member.String(), row.Username)
		if row.ID == known.ID {
			require.NotNil(t, row.BrowserName)
			require.Equal(t, chrome, *row.BrowserName)
		} else {
			require.Equal(t, unknown.ID, row.ID)
			require.Nil(t, row.BrowserName)
		}
	}
	// An original issuance value can fill a row tracked by an older API.
	unknown.BrowserName = &safari
	require.NoError(t, f.repo.CheckSession(t.Context(), member, unknown))
	rows, err = f.repo.Sessions(t.Context(), f.scope, &f.member, false)
	require.NoError(t, err)
	for _, row := range rows.Items {
		if row.ID == unknown.ID {
			require.NotNil(t, row.BrowserName)
			require.Equal(t, safari, *row.BrowserName)
		}
	}
}
