package maya

import (
	"context"
	"fmt"
	"strings"
	"time"
	"unicode/utf8"

	mayadomain "github.com/complexus-tech/projects-api/internal/modules/maya/domain"
	"github.com/google/uuid"
)

type Skill = mayadomain.Skill
type SkillScope = mayadomain.SkillScope
type SkillContent = mayadomain.SkillContent

var (
	ErrSkillNotFound  = mayadomain.ErrSkillNotFound
	ErrInvalidSkill   = mayadomain.ErrInvalidSkill
	ErrSkillNameTaken = mayadomain.ErrSkillNameTaken
	ErrSkillChanged   = mayadomain.ErrSkillChanged
)

type SkillsRepository interface {
	ListSkills(context.Context, SkillScope) ([]Skill, error)
	CreateSkill(context.Context, SkillScope, SkillContent) (Skill, error)
	UpdateSkill(context.Context, SkillScope, uuid.UUID, SkillContent, time.Time) (Skill, error)
	DeleteSkill(context.Context, SkillScope, uuid.UUID) error
}

func (s *Service) skillsRepository(scope SkillScope) (SkillsRepository, error) {
	if scope.WorkspaceID == uuid.Nil || scope.UserID == uuid.Nil {
		return nil, fmt.Errorf("%w: workspace and user are required", ErrInvalidSkill)
	}
	if s == nil {
		return nil, ErrNotConfigured
	}
	repo, ok := s.repo.(SkillsRepository)
	if !ok || repo == nil {
		return nil, ErrNotConfigured
	}
	return repo, nil
}

func validateSkillContent(content SkillContent) (SkillContent, error) {
	content.Name = strings.TrimSpace(content.Name)
	content.Description = strings.TrimSpace(content.Description)
	content.Instructions = strings.TrimSpace(content.Instructions)
	if length := utf8.RuneCountInString(content.Name); length < 1 || length > 80 {
		return SkillContent{}, fmt.Errorf("%w: name must be between 1 and 80 characters", ErrInvalidSkill)
	}
	if utf8.RuneCountInString(content.Description) > 300 {
		return SkillContent{}, fmt.Errorf("%w: description must be 300 characters or fewer", ErrInvalidSkill)
	}
	if length := utf8.RuneCountInString(content.Instructions); length < 1 || length > 12000 {
		return SkillContent{}, fmt.Errorf("%w: instructions must be between 1 and 12000 characters", ErrInvalidSkill)
	}
	return content, nil
}

func (s *Service) ListSkills(ctx context.Context, scope SkillScope) ([]Skill, error) {
	repo, err := s.skillsRepository(scope)
	if err != nil {
		return nil, err
	}
	return repo.ListSkills(ctx, scope)
}

func (s *Service) CreateSkill(ctx context.Context, scope SkillScope, content SkillContent) (Skill, error) {
	repo, err := s.skillsRepository(scope)
	if err != nil {
		return Skill{}, err
	}
	content, err = validateSkillContent(content)
	if err != nil {
		return Skill{}, err
	}
	return repo.CreateSkill(ctx, scope, content)
}

func (s *Service) UpdateSkill(ctx context.Context, scope SkillScope, id uuid.UUID, content SkillContent, updatedAt time.Time) (Skill, error) {
	repo, err := s.skillsRepository(scope)
	if err != nil {
		return Skill{}, err
	}
	if id == uuid.Nil || updatedAt.IsZero() {
		return Skill{}, fmt.Errorf("%w: skill and previous update time are required", ErrInvalidSkill)
	}
	content, err = validateSkillContent(content)
	if err != nil {
		return Skill{}, err
	}
	return repo.UpdateSkill(ctx, scope, id, content, updatedAt)
}

func (s *Service) DeleteSkill(ctx context.Context, scope SkillScope, id uuid.UUID) error {
	repo, err := s.skillsRepository(scope)
	if err != nil {
		return err
	}
	if id == uuid.Nil {
		return fmt.Errorf("%w: skill is required", ErrInvalidSkill)
	}
	return repo.DeleteSkill(ctx, scope, id)
}
