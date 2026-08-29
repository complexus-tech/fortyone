package codehost

import (
	"errors"
	"testing"

	"github.com/complexus-tech/projects-api/internal/platform/integrations"
	"github.com/google/uuid"
)

func TestCapabilitySetMakesUnsupportedBehaviorExplicit(t *testing.T) {
	t.Parallel()
	capabilities := Capabilities{CapabilityRepositoryCatalog: true}
	if err := capabilities.Require(CapabilityRepositoryCatalog); err != nil {
		t.Fatalf("require repository catalog: %v", err)
	}
	if err := capabilities.Require(CapabilityCommentWriter); !errors.Is(err, ErrCapabilityUnsupported) {
		t.Fatalf("unsupported capability error = %v", err)
	}
}

func TestInstallationAndCursorValidation(t *testing.T) {
	t.Parallel()
	installation := InstallationRef{
		Provider:               integrations.ProviderKey("gitlab"),
		WorkspaceID:            uuid.New(),
		InstallationID:         uuid.New(),
		ExternalInstallationID: "42",
		Generation:             uuid.New(),
	}
	if err := ValidateInstallation(installation); err != nil {
		t.Fatalf("validate installation: %v", err)
	}
	if err := ValidateCursor(Cursor{Limit: 100}); err != nil {
		t.Fatalf("validate cursor: %v", err)
	}
	if err := ValidateCursor(Cursor{Limit: 101}); !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("oversized cursor error = %v", err)
	}
}

func TestCommentCommandDerivesRepositoryFromWorkItem(t *testing.T) {
	t.Parallel()
	command := AddComment{
		WorkItem: WorkItem{
			Number: 7,
			Repository: RepositoryRef{
				ExternalID: "42",
				Owner:      "complexus",
				Name:       "fortyone",
				FullName:   "complexus/fortyone",
				WebURL:     "https://gitlab.example/complexus/fortyone",
			},
		},
		Body: "one repository identity",
	}
	if err := ValidateAddComment(command); err != nil {
		t.Fatalf("ValidateAddComment() error = %v", err)
	}
	command.WorkItem.Repository.WebURL = "http://gitlab.example/complexus/fortyone"
	if err := ValidateAddComment(command); !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("ValidateAddComment(insecure URL) error = %v, want ErrInvalidInput", err)
	}
	command.WorkItem.Repository.WebURL = "https://gitlab.example/complexus/fortyone"
	command.WorkItem.Repository = RepositoryRef{}
	if err := ValidateAddComment(command); !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("ValidateAddComment(missing repository) error = %v, want ErrInvalidInput", err)
	}
}
