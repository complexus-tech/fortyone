package cache

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

const (
	// TTLs
	DefaultTTL    = 10 * time.Minute
	ListTTL       = 10 * time.Minute
	DetailTTL     = 15 * time.Minute
	ShortTermTTL  = 3 * time.Minute
	MediumTermTTL = 30 * time.Minute
	LongTermTTL   = 2 * time.Hour
)

// Key formats for different types of cache entries
const (
	ObjectiveListKey     = "objectives:list:%s"        // workspaceID
	ObjectiveDetailKey   = "objectives:detail:%s:%s"   // workspaceID, objectiveID
	KeyResultsListKey    = "key-results:list:%s:%s"    // workspaceID, objectiveID
	StoryListKey         = "stories:list:%s"           // workspaceID
	StoryDetailKey       = "stories:detail:%s:%s"      // workspaceID, storyID
	WorkspaceDetailKey   = "workspaces:detail:%s"      // workspaceID
	UserDetailKey        = "users:detail:%s"           // userID
	MyStoriesKey         = "stories:my-stories:%s"     // workspaceID
	StoryCommentsKey     = "stories:comments:%s:%s"    // workspaceID, storyID
	StoryActivitiesKey   = "stories:activities:%s:%s"  // workspaceID, storyID
	StoryAttachmentsKey  = "stories:attachments:%s:%s" // workspaceID, storyID
	WorkspacesListKey    = "workspaces:list:%s"        // userID (list of workspaces for a user)
	WorkspaceMembersKey  = "workspaces:members:%s"     // workspaceID (members of a workspace)
	WorkspaceTeamsKey    = "workspaces:teams:%s"       // workspaceID (teams in a workspace)
	WorkspaceSettingsKey = "workspaces:settings:%s"    // workspaceID
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

// StoryCommentsCacheKey generates a cache key for comments of a story
func StoryCommentsCacheKey(workspaceID, storyID uuid.UUID) string {
	return fmt.Sprintf(StoryCommentsKey, workspaceID.String(), storyID.String())
}

// StoryActivitiesCacheKey generates a cache key for activities of a story
func StoryActivitiesCacheKey(workspaceID, storyID uuid.UUID) string {
	return fmt.Sprintf(StoryActivitiesKey, workspaceID.String(), storyID.String())
}

// StoryAttachmentsCacheKey generates a cache key for attachments of a story
func StoryAttachmentsCacheKey(workspaceID, storyID uuid.UUID) string {
	return fmt.Sprintf(StoryAttachmentsKey, workspaceID.String(), storyID.String())
}

// WorkspacesListCacheKey generates a cache key for a user's list of workspaces
func WorkspacesListCacheKey(userID uuid.UUID) string {
	return fmt.Sprintf(WorkspacesListKey, userID.String())
}

// WorkspaceMembersCacheKey generates a cache key for members of a workspace
func WorkspaceMembersCacheKey(workspaceID uuid.UUID) string {
	return fmt.Sprintf(WorkspaceMembersKey, workspaceID.String())
}

// WorkspaceTeamsCacheKey generates a cache key for teams in a workspace
func WorkspaceTeamsCacheKey(workspaceID uuid.UUID) string {
	return fmt.Sprintf(WorkspaceTeamsKey, workspaceID.String())
}

// WorkspaceSettingsCacheKey generates a cache key for workspace settings
func WorkspaceSettingsCacheKey(workspaceID uuid.UUID) string {
	return fmt.Sprintf(WorkspaceSettingsKey, workspaceID.String())
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
		StoryCommentsCacheKey(workspaceID, storyID),
		StoryActivitiesCacheKey(workspaceID, storyID),
		StoryAttachmentsCacheKey(workspaceID, storyID),
	}
}

// InvalidateWorkspaceKeys invalidates all cache keys related to a workspace
func InvalidateWorkspaceKeys(workspaceID uuid.UUID) []string {
	return []string{
		WorkspaceDetailCacheKey(workspaceID),
		WorkspaceMembersCacheKey(workspaceID),
		WorkspaceTeamsCacheKey(workspaceID),
		WorkspaceSettingsCacheKey(workspaceID),
		fmt.Sprintf(ObjectiveListKey+"*", workspaceID.String()),
		fmt.Sprintf(StoryListKey+"*", workspaceID.String()),
	}
}

// InvalidateKeyResultKeys invalidates all cache keys related to key results and their parent objectives
func InvalidateKeyResultKeys(workspaceID uuid.UUID) []string {
	return []string{
		fmt.Sprintf(KeyResultsListKey+"*", workspaceID.String(), "*"), // Wildcard to match all key results lists for any objective
		fmt.Sprintf(ObjectiveListKey+"*", workspaceID.String()),       // Objective lists may show key result data
	}
}

// InvalidateUserWorkspacesKeys invalidates cache keys for a user's workspace list
func InvalidateUserWorkspacesKeys(userID uuid.UUID) []string {
	return []string{
		WorkspacesListCacheKey(userID),
	}
}
