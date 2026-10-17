package chatdomain

import "github.com/google/uuid"

// MemberSearch searches the full current directory without expanding avatar payloads.
type MemberSearch struct {
	Scope          ActorScope
	ConversationID uuid.UUID
	Query          string
	Limit          int
}

type MemberPage struct {
	Items      []Person `json:"items"`
	NextCursor *string  `json:"nextCursor"`
}
