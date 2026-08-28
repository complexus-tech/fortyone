package workspacedomain

import "errors"

var (
	ErrNotFound               = errors.New("workspace not found")
	ErrMemberNotFound         = errors.New("member not found")
	ErrSlugTaken              = errors.New("workspace with this url already exists")
	ErrRestrictedSlug         = errors.New("this workspace slug is restricted")
	ErrAlreadyWorkspaceMember = errors.New("user is already a member of this workspace")
)
