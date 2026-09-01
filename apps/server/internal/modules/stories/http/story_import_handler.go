package storieshttp

import (
	"context"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"hash"
	"net/http"
	"strings"

	storydomain "github.com/complexus-tech/projects-api/internal/modules/stories/domain"
	stories "github.com/complexus-tech/projects-api/internal/modules/stories/service"
	mid "github.com/complexus-tech/projects-api/internal/platform/http/middleware"
	"github.com/complexus-tech/projects-api/pkg/cache"
	"github.com/complexus-tech/projects-api/pkg/web"
	"github.com/google/uuid"
)

func (h *Handlers) Import(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	ctx, span := web.AddSpan(ctx, "storieshttp.handlers.Import")
	defer span.End()
	w.Header().Set("Cache-Control", "private, no-store")

	workspace, err := mid.GetWorkspace(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}
	actorID, err := mid.GetUserID(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}
	if h.storyImporter == nil {
		return web.RespondError(ctx, w, errors.New("story import service is unavailable"), http.StatusServiceUnavailable)
	}

	var request AppStoryImportRequest
	if err := web.Decode(r, &request); err != nil {
		return web.RespondError(ctx, w, err, http.StatusBadRequest)
	}

	provider := strings.ToLower(request.Provider)
	sourceDigest := strings.ToLower(request.SourceDigest)
	response := AppStoryImportResponse{
		Counts: AppStoryImportCounts{Total: len(request.Items)},
		Items:  make([]AppStoryImportItemResult, 0, len(request.Items)),
	}
	for _, item := range request.Items {
		newStory := toCoreNewStory(item.Story, actorID)
		creationKey := storyImportCreationKey(
			workspace.ID,
			newStory.Team,
			provider,
			sourceDigest,
			request.SourceNamespace,
			item.SourceKey,
		)
		newStory.CreationKey = &creationKey
		newStory.ExternalDelivery = storydomain.ExternalStoryDeliveryInternalOnly

		createdStory, createErr := h.storyImporter.CreateExternal(ctx, actorID, newStory, workspace.ID)
		result := AppStoryImportItemResult{SourceKey: item.SourceKey}
		if createErr != nil {
			result.Error = storyImportItemError(createErr)
			response.Counts.Failed++
			response.Items = append(response.Items, result)
			continue
		}

		storyID := createdStory.ID
		result.StoryID = &storyID
		result.Created = createdStory.CreatedNow
		if createdStory.CreatedNow {
			response.Counts.Created++
		} else {
			response.Counts.Replayed++
		}
		response.Items = append(response.Items, result)
	}

	if response.Counts.Created+response.Counts.Replayed > 0 && h.cache != nil {
		h.cache.DeleteByPattern(ctx, fmt.Sprintf(cache.StoryListKey+"*", workspace.ID.String()))
		h.cache.DeleteByPattern(ctx, fmt.Sprintf(cache.MyStoriesKey+"*", workspace.ID.String()))
	}

	return web.Respond(ctx, w, response, http.StatusOK)
}

func storyImportCreationKey(
	workspaceID, teamID uuid.UUID,
	provider, sourceDigest string,
	sourceNamespace *string,
	sourceKey string,
) string {
	provider = strings.ToLower(provider)
	sourceDigest = strings.ToLower(sourceDigest)
	digest := sha256.New()
	writeImportCreationKeyPart(digest, "story-import-v1")
	writeImportCreationKeyPart(digest, workspaceID.String())
	writeImportCreationKeyPart(digest, teamID.String())
	writeImportCreationKeyPart(digest, provider)
	if sourceNamespace != nil {
		writeImportCreationKeyPart(digest, "source-namespace")
		writeImportCreationKeyPart(digest, *sourceNamespace)
	} else if provider == storyImportProviderJiraCSV {
		// Jira issue keys are stable across refreshed exports, while the file
		// digest is not. CSV does not carry a trustworthy Jira Cloud/site ID,
		// so this intentionally scopes the normalized issue key only to the
		// destination workspace and team. See the importer plan for that
		// limitation.
		sourceKey = strings.ToUpper(sourceKey)
	} else {
		writeImportCreationKeyPart(digest, sourceDigest)
	}
	writeImportCreationKeyPart(digest, sourceKey)
	return "story-import:v1:" + hex.EncodeToString(digest.Sum(nil))
}

func writeImportCreationKeyPart(digest hash.Hash, value string) {
	var length [4]byte
	binary.BigEndian.PutUint32(length[:], uint32(len(value)))
	_, _ = digest.Write(length[:])
	_, _ = digest.Write([]byte(value))
}

func storyImportItemError(err error) *AppStoryImportItemError {
	switch storyMutationStatus(err) {
	case http.StatusBadRequest:
		return &AppStoryImportItemError{Code: "invalid_story", Message: "The reviewed story mapping is invalid."}
	case http.StatusForbidden:
		return &AppStoryImportItemError{Code: "permission_denied", Message: "The story cannot be created with the selected mappings."}
	case http.StatusNotFound:
		return &AppStoryImportItemError{Code: "mapped_resource_not_found", Message: "A selected destination resource no longer exists."}
	case http.StatusConflict:
		return &AppStoryImportItemError{Code: "conflict", Message: "The story conflicts with the current workspace state."}
	case http.StatusPaymentRequired:
		return &AppStoryImportItemError{Code: "feature_unavailable", Message: "A selected story feature is not available for this workspace."}
	case http.StatusServiceUnavailable:
		return &AppStoryImportItemError{Code: "service_unavailable", Message: "The story could not be imported because a required service is unavailable."}
	default:
		return &AppStoryImportItemError{Code: "internal_error", Message: "The story could not be imported."}
	}
}

var _ storyImportService = (*stories.Service)(nil)
