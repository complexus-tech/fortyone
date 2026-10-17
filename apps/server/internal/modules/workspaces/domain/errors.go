package workspacedomain

import "errors"

var (
	ErrSettingsConflict       = errors.New("workspace settings changed; refresh and try again")
	ErrNotFound               = errors.New("workspace not found")
	ErrMemberNotFound         = errors.New("member not found")
	ErrSlugTaken              = errors.New("workspace with this url already exists")
	ErrRestrictedSlug         = errors.New("this workspace slug is restricted")
	ErrAlreadyWorkspaceMember = errors.New("user is already a member of this workspace")
)
