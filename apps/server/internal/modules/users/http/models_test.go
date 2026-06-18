package usershttp

import (
	"testing"

	users "github.com/complexus-tech/projects-api/internal/modules/users/service"
)

func TestToAppUserIncludesInternalFlag(t *testing.T) {
	appUser := toAppUser(users.CoreUser{
		IsInternal: true,
	})

	if !appUser.IsInternal {
		t.Fatal("expected app user to expose internal flag")
	}
}
