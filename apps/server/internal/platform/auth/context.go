package auth

import (
	"context"

	"github.com/google/uuid"
)

type contextKey string

const actorKey contextKey = "actor"

// SetUserID is the compatibility entry point for first-party human sessions.
// New credential types should construct an Actor explicitly with SetActor.
func SetUserID(ctx context.Context, userID uuid.UUID) context.Context {
	return context.WithValue(ctx, actorKey, NewHumanActor(userID))
}

func GetUserID(ctx context.Context) (uuid.UUID, error) {
	actor, err := GetActor(ctx)
	if err != nil {
		return uuid.Nil, err
	}
	return actor.UserID()
}

func SetActor(ctx context.Context, actor Actor) (context.Context, error) {
	if err := actor.Validate(); err != nil {
		return nil, err
	}
	return context.WithValue(ctx, actorKey, actor.clone()), nil
}

func GetActor(ctx context.Context) (Actor, error) {
	actor, ok := ctx.Value(actorKey).(Actor)
	if !ok {
		return Actor{}, ErrActorNotFound
	}
	if err := actor.Validate(); err != nil {
		return Actor{}, err
	}
	return actor.clone(), nil
}

func BindWorkspace(ctx context.Context, workspaceID uuid.UUID) (context.Context, error) {
	actor, err := GetActor(ctx)
	if err != nil {
		return nil, err
	}
	actor, err = actor.WithWorkspace(workspaceID)
	if err != nil {
		return nil, err
	}
	return SetActor(ctx, actor)
}
