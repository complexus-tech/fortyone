package feedbackhttp

import "github.com/google/uuid"

// This is the migration-reserved identity for retained workspace contributions,
// never a participant profile. Keep it distinct from ordinary anonymous guests.
const formerUserID = "ffffffff-ffff-4fff-8fff-ffffffffffff"

func presentAuthor(id uuid.UUID, name string, avatar *string, masked bool) (*uuid.UUID, string, *string) {
	if id.String() == formerUserID {
		return nil, "Former user", nil
	}
	if masked {
		return nil, name, nil
	}
	return uuidPointer(id), name, avatar
}
