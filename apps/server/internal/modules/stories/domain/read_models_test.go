package domain

import (
	"errors"
	"testing"

	"github.com/google/uuid"
)

func TestReadScopeValidateRejectsAmbiguousOrMissingAuthority(t *testing.T) {
	actorID := uuid.New()
	workspaceID := uuid.New()
	teamID := uuid.New()

	tests := []struct {
		name  string
		scope ReadScope
	}{
		{name: "missing actor", scope: ReadScope{WorkspaceID: workspaceID, UnrestrictedTeamAccess: true}},
		{name: "missing workspace", scope: ReadScope{ActorID: actorID, UnrestrictedTeamAccess: true}},
		{name: "empty restriction", scope: ReadScope{ActorID: actorID, WorkspaceID: workspaceID}},
		{name: "zero restricted team", scope: ReadScope{ActorID: actorID, WorkspaceID: workspaceID, AllowedTeamIDs: []uuid.UUID{uuid.Nil}}},
		{name: "ambiguous restriction", scope: ReadScope{ActorID: actorID, WorkspaceID: workspaceID, UnrestrictedTeamAccess: true, AllowedTeamIDs: []uuid.UUID{teamID}}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.scope.Validate(); !errors.Is(err, ErrInvalidReadScope) {
				t.Fatalf("error = %v, want ErrInvalidReadScope", err)
			}
		})
	}
}

func TestReadScopeValidateAcceptsExplicitTeamModes(t *testing.T) {
	base := ReadScope{ActorID: uuid.New(), WorkspaceID: uuid.New()}
	for _, scope := range []ReadScope{
		{ActorID: base.ActorID, WorkspaceID: base.WorkspaceID, UnrestrictedTeamAccess: true},
		{ActorID: base.ActorID, WorkspaceID: base.WorkspaceID, AllowedTeamIDs: []uuid.UUID{uuid.New()}},
	} {
		if err := scope.Validate(); err != nil {
			t.Fatalf("valid scope rejected: %v", err)
		}
	}
}
