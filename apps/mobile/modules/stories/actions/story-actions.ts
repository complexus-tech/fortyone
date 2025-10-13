import { post, remove } from "@/lib/http";
import type { ApiResponse } from "@/types";

type BulkPayload = {
  storyIds: string[];
};

// Single story endpoints
export const duplicateStory = async (storyId: string) => {
  const response = await post<null, ApiResponse<{ id: string }>>(
    `stories/${storyId}/duplicate`,
    null
  );
  return response;
};

export const restoreStory = async (storyId: string) => {
  const response = await post<null, ApiResponse<null>>(
    `stories/${storyId}/restore`,
    null
  );
  return response;
};

export const deleteStory = async (storyId: string) => {
  const response = await remove<ApiResponse<null>>(`stories/${storyId}`);
  return response;
};

// Bulk endpoints (used for single stories too)
export const archiveStory = async (storyIds: string[]) => {
  const response = await post<BulkPayload, ApiResponse<BulkPayload>>(
    "stories/archive",
    { storyIds }
  );
  return response;
};

export const unarchiveStory = async (storyIds: string[]) => {
  const response = await post<BulkPayload, ApiResponse<BulkPayload>>(
    "stories/unarchive",
    { storyIds }
  );
  return response;
};
