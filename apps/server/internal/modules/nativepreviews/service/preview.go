package nativepreviews

import (
	"context"
	"crypto/sha256"
	"fmt"
	"time"

	domain "github.com/complexus-tech/projects-api/internal/modules/nativepreviews/domain"
	"github.com/complexus-tech/projects-api/internal/platform/auth"
	"github.com/google/uuid"
)

type Repository interface {
	Read(context.Context, auth.Actor, uuid.UUID, domain.Target, bool) (domain.Snapshot, error)
}

type Cache interface {
	Get(context.Context, string, any) error
	Set(context.Context, string, any, time.Duration) error
}

type Service struct {
	repository Repository
	cache      Cache
}

func New(repository Repository, cache Cache) *Service {
	return &Service{repository: repository, cache: cache}
}

func (s *Service) Get(ctx context.Context, workspaceID uuid.UUID, target domain.Target, refresh bool) (domain.Preview, error) {
	actor, err := auth.GetActor(ctx)
	if err != nil || actor.WorkspaceID != workspaceID || !actor.IsUserActor() {
		return domain.Preview{}, domain.ErrForbidden
	}
	var required auth.Scope
	switch target.Kind {
	case "story":
		required = auth.ScopeStoriesRead
	case "objective":
		required = auth.ScopeObjectivesRead
	case "document":
		required = auth.ScopeFirstParty
	default:
		return domain.Preview{}, domain.ErrInvalidTarget
	}
	if !actor.Scopes.Has(required) {
		return domain.Preview{}, domain.ErrForbidden
	}

	// Authorization is deliberately never cached, even when content is already
	// in Redis. This also checks current existence, membership and visibility.
	snapshot, err := s.repository.Read(ctx, actor, workspaceID, target, refresh)
	if err != nil {
		return domain.Preview{}, err
	}
	key := cacheKey(workspaceID, target.Kind, snapshot)
	if !refresh && s.cache != nil {
		var cached domain.Preview
		cacheCtx, cancel := context.WithTimeout(ctx, 100*time.Millisecond)
		err := s.cache.Get(cacheCtx, key, &cached)
		cancel()
		if err == nil && cached.Kind == target.Kind && cached.Version == snapshot.Version && len(cached.Value) > 0 {
			return cached, nil
		}
	}
	if !refresh {
		snapshot, err = s.repository.Read(ctx, actor, workspaceID, target, true)
		if err != nil {
			return domain.Preview{}, err
		}
		key = cacheKey(workspaceID, target.Kind, snapshot)
	}
	preview := domain.Preview{Kind: target.Kind, Value: snapshot.Value, Version: snapshot.Version, RefreshedAt: time.Now().UTC()}
	if s.cache != nil {
		cacheCtx, cancel := context.WithTimeout(ctx, 100*time.Millisecond)
		// Redis is an optional acceleration layer; a failed write must not hide
		// an authorized, freshly loaded preview. The cache service logs failures.
		_ = s.cache.Set(cacheCtx, key, preview, time.Hour)
		cancel()
	}
	return preview, nil
}

func cacheKey(workspaceID uuid.UUID, kind string, snapshot domain.Snapshot) string {
	version := sha256.Sum256([]byte(snapshot.Version))
	return fmt.Sprintf("native-preview:v1:%s:%s:%s:%x", workspaceID, kind, snapshot.ID, version[:16])
}
