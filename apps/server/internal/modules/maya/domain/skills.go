package mayadomain

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrSkillNotFound  = errors.New("skill not found")
	ErrInvalidSkill   = errors.New("invalid skill")
	ErrSkillNameTaken = errors.New("a skill with this name already exists")
	ErrSkillChanged   = errors.New("this skill changed; reopen it to edit the latest version")
)

type Skill struct {
	ID           uuid.UUID
	Name         string
	Description  string
	Instructions string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type SkillScope struct {
	WorkspaceID uuid.UUID
	UserID      uuid.UUID
}

type SkillContent struct {
	Name         string
	Description  string
	Instructions string
}
