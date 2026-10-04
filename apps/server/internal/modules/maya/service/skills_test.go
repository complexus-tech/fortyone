package maya

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestSkillsValidateContentBeforePersistence(t *testing.T) {
	t.Parallel()
	for _, test := range []struct {
		name    string
		content SkillContent
	}{
		{name: "blank name", content: SkillContent{Name: " \n ", Instructions: "Triage"}},
		{name: "long name", content: SkillContent{Name: strings.Repeat("a", 81), Instructions: "Triage"}},
		{name: "long description", content: SkillContent{Name: "Triage", Description: strings.Repeat("a", 301), Instructions: "Triage"}},
		{name: "blank instructions", content: SkillContent{Name: "Triage", Instructions: " \n "}},
		{name: "long instructions", content: SkillContent{Name: "Triage", Instructions: strings.Repeat("a", 12001)}},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			repo := &skillsRepositoryStub{}
			service := New(Dependencies{Repository: repo})
			_, err := service.CreateSkill(t.Context(), SkillScope{WorkspaceID: uuid.New(), UserID: uuid.New()}, test.content)
			if !errors.Is(err, ErrInvalidSkill) || repo.created {
				t.Fatalf("create error = %v, persistence called = %t", err, repo.created)
			}
		})
	}
}

func TestSkillsTrimContentAndPreserveOwnerAndVersion(t *testing.T) {
	t.Parallel()
	repo := &skillsRepositoryStub{}
	service := New(Dependencies{Repository: repo})
	scope := SkillScope{WorkspaceID: uuid.New(), UserID: uuid.New()}
	content := SkillContent{Name: "  Triage  ", Description: "  My inbox  ", Instructions: "  Review tasks.\nAsk before edits.  "}
	_, err := service.CreateSkill(t.Context(), scope, content)
	if err != nil || repo.scope != scope || repo.content.Name != "Triage" || repo.content.Description != "My inbox" || repo.content.Instructions != "Review tasks.\nAsk before edits." {
		t.Fatalf("create = %v, scope = %+v, content = %+v", err, repo.scope, repo.content)
	}
	id := uuid.New()
	version := time.Now().UTC()
	_, err = service.UpdateSkill(t.Context(), scope, id, content, version)
	if err != nil || repo.id != id || !repo.version.Equal(version) {
		t.Fatalf("update = %v, id = %v, version = %v", err, repo.id, repo.version)
	}
}

func TestSkillsRejectMissingScopeAndEditVersion(t *testing.T) {
	t.Parallel()
	repo := &skillsRepositoryStub{}
	service := New(Dependencies{Repository: repo})
	if _, err := service.ListSkills(t.Context(), SkillScope{}); !errors.Is(err, ErrInvalidSkill) {
		t.Fatalf("unscoped list error = %v", err)
	}
	scope := SkillScope{WorkspaceID: uuid.New(), UserID: uuid.New()}
	if _, err := service.UpdateSkill(t.Context(), scope, uuid.New(), SkillContent{Name: "Triage", Instructions: "Review"}, time.Time{}); !errors.Is(err, ErrInvalidSkill) || repo.updated {
		t.Fatalf("unversioned edit error = %v, persisted = %t", err, repo.updated)
	}
}

type skillsRepositoryStub struct {
	Repository
	scope   SkillScope
	content SkillContent
	id      uuid.UUID
	version time.Time
	created bool
	updated bool
}

func (repo *skillsRepositoryStub) ListSkills(_ context.Context, scope SkillScope) ([]Skill, error) {
	repo.scope = scope
	return []Skill{}, nil
}

func (repo *skillsRepositoryStub) CreateSkill(_ context.Context, scope SkillScope, content SkillContent) (Skill, error) {
	repo.scope, repo.content, repo.created = scope, content, true
	return Skill{}, nil
}

func (repo *skillsRepositoryStub) UpdateSkill(_ context.Context, scope SkillScope, id uuid.UUID, content SkillContent, version time.Time) (Skill, error) {
	repo.scope, repo.content, repo.id, repo.version, repo.updated = scope, content, id, version, true
	return Skill{}, nil
}

func (repo *skillsRepositoryStub) DeleteSkill(_ context.Context, scope SkillScope, id uuid.UUID) error {
	repo.scope, repo.id = scope, id
	return nil
}
