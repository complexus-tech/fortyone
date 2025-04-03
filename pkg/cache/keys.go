package cache

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

const (
	// TTLs
	DefaultTTL    = 5 * time.Minute
	ListTTL       = 2 * time.Minute
	DetailTTL     = 5 * time.Minute
	ShortTermTTL  = 1 * time.Minute
	MediumTermTTL = 15 * time.Minute
	LongTermTTL   = 1 * time.Hour
)

// Key formats for different types of cache entries
const (
	ObjectiveListKey   = "objectives:list:%s"      // workspaceID
	ObjectiveDetailKey = "objectives:detail:%s:%s" // workspaceID, objectiveID
	KeyResultsListKey  = "key-results:list:%s:%s"  // workspaceID, objectiveID
	StoryListKey       = "stories:list:%s"         // workspaceID
	StoryDetailKey     = "stories:detail:%s:%s"    // workspaceID, storyID
	WorkspaceDetailKey = "workspaces:detail:%s"    // workspaceID
	UserDetailKey      = "users:detail:%s"         // userID
)

// ObjectiveListCacheKey generates a cache key for a list of objectives
func ObjectiveListCacheKey(workspaceID uuid.UUID, filters string) string {
	if filters == "" {
		return fmt.Sprintf(ObjectiveListKey, workspaceID.String())
	}
	return fmt.Sprintf(ObjectiveListKey+":%s", workspaceID.String(), filters)
}

// ObjectiveDetailCacheKey generates a cache key for a single objective
func ObjectiveDetailCacheKey(workspaceID, objectiveID uuid.UUID) string {
	return fmt.Sprintf(ObjectiveDetailKey, workspaceID.String(), objectiveID.String())
}

// KeyResultsListCacheKey generates a cache key for a list of key results for an objective
func KeyResultsListCacheKey(workspaceID, objectiveID uuid.UUID) string {
	return fmt.Sprintf(KeyResultsListKey, workspaceID.String(), objectiveID.String())
}

// StoryListCacheKey generates a cache key for a list of stories
func StoryListCacheKey(workspaceID uuid.UUID, filters string) string {
	if filters == "" {
		return fmt.Sprintf(StoryListKey, workspaceID.String())
	}
	return fmt.Sprintf(StoryListKey+":%s", workspaceID.String(), filters)
}

// StoryDetailCacheKey generates a cache key for a single story
func StoryDetailCacheKey(workspaceID, storyID uuid.UUID) string {
	return fmt.Sprintf(StoryDetailKey, workspaceID.String(), storyID.String())
}

// WorkspaceDetailCacheKey generates a cache key for a single workspace
func WorkspaceDetailCacheKey(workspaceID uuid.UUID) string {
	return fmt.Sprintf(WorkspaceDetailKey, workspaceID.String())
}

// UserDetailCacheKey generates a cache key for a single user
func UserDetailCacheKey(userID uuid.UUID) string {
	return fmt.Sprintf(UserDetailKey, userID.String())
}

// InvalidateObjectiveKeys invalidates all cache keys related to an objective
func InvalidateObjectiveKeys(workspaceID, objectiveID uuid.UUID) []string {
	return []string{
		ObjectiveDetailCacheKey(workspaceID, objectiveID),
		fmt.Sprintf(ObjectiveListKey+"*", workspaceID.String()), // Wildcard to match all list variations
		KeyResultsListCacheKey(workspaceID, objectiveID),
	}
}

// InvalidateStoryKeys invalidates all cache keys related to a story
func InvalidateStoryKeys(workspaceID, storyID uuid.UUID) []string {
	return []string{
		StoryDetailCacheKey(workspaceID, storyID),
		fmt.Sprintf(StoryListKey+"*", workspaceID.String()), // Wildcard to match all list variations
	}
}

// InvalidateWorkspaceKeys invalidates all cache keys related to a workspace
func InvalidateWorkspaceKeys(workspaceID uuid.UUID) []string {
	return []string{
		WorkspaceDetailCacheKey(workspaceID),
		fmt.Sprintf(ObjectiveListKey+"*", workspaceID.String()),
		fmt.Sprintf(StoryListKey+"*", workspaceID.String()),
	}
}
