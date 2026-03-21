package actors

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/complexus-tech/projects-api/pkg/cache"
	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type Key string

const (
	KeySystem Key = "system"
	KeyGitHub Key = "github"
)

const cacheTTL = time.Hour

var actorEmails = map[Key]string{
	KeySystem: "maya@fortyone.app",
	KeyGitHub: "github@fortyone.app",
}

type Resolver struct {
	log   *logger.Logger
	db    *sqlx.DB
	cache *cache.Service
}

func NewResolver(log *logger.Logger, db *sqlx.DB, cache *cache.Service) *Resolver {
	return &Resolver{
		log:   log,
		db:    db,
		cache: cache,
	}
}

func (r *Resolver) Resolve(ctx context.Context, key Key) (uuid.UUID, error) {
	email, ok := actorEmails[key]
	if !ok {
		return uuid.Nil, fmt.Errorf("unknown actor key: %s", key)
	}

	if actorID, ok, err := r.getCached(ctx, key); err != nil {
		r.log.Warn(ctx, "failed to resolve actor from cache", "actor_key", key, "error", err)
	} else if ok {
		return actorID, nil
	}

	actorID, err := r.getFromDB(ctx, email)
	if err != nil {
		return uuid.Nil, err
	}

	if r.cache != nil {
		if err := r.cache.Set(ctx, cacheKey(key), actorID.String(), cacheTTL); err != nil {
			r.log.Warn(ctx, "failed to cache actor id", "actor_key", key, "error", err)
		}
	}

	return actorID, nil
}

func (r *Resolver) getCached(ctx context.Context, key Key) (uuid.UUID, bool, error) {
	if r.cache == nil {
		return uuid.Nil, false, nil
	}

	var actorID string
	if err := r.cache.Get(ctx, cacheKey(key), &actorID); err != nil {
		if errors.Is(err, cache.ErrNotFound) {
			return uuid.Nil, false, nil
		}
		return uuid.Nil, false, err
	}

	parsedID, err := uuid.Parse(actorID)
	if err != nil {
		if deleteErr := r.cache.Delete(ctx, cacheKey(key)); deleteErr != nil {
			r.log.Warn(ctx, "failed to delete invalid cached actor id", "actor_key", key, "error", deleteErr)
		}
		return uuid.Nil, false, fmt.Errorf("invalid cached actor id for %s: %w", key, err)
	}

	return parsedID, true, nil
}

func (r *Resolver) getFromDB(ctx context.Context, email string) (uuid.UUID, error) {
	const q = `
		SELECT user_id
		FROM users
		WHERE email = $1
			AND is_system = true
		LIMIT 1
	`

	var actorID uuid.UUID
	if err := r.db.GetContext(ctx, &actorID, q, email); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return uuid.Nil, fmt.Errorf("system actor not found for email %q", email)
		}
		return uuid.Nil, fmt.Errorf("failed to resolve system actor %q: %w", email, err)
	}

	return actorID, nil
}

func cacheKey(key Key) string {
	return fmt.Sprintf("actors:user-id:%s", key)
}
