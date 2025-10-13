import { post, remove } from "@/lib/http";
import type { ApiResponse } from "@/types";

export const archiveStory = async (storyId: string) => {
  const response = await post<null, ApiResponse<null>>(
    `stories/${storyId}/archive`,
    null
  );
  return response;
};

export const unarchiveStory = async (storyId: string) => {
  const response = await post<null, ApiResponse<null>>(
    `stories/${storyId}/unarchive`,
    null
  );
  return response;
};

export const deleteStory = async (storyId: string, hardDelete = false) => {
  const response = await remove<ApiResponse<null>>(
    `stories/${storyId}${hardDelete ? "?hard=true" : ""}`
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

export const duplicateStory = async (storyId: string) => {
  const response = await post<null, ApiResponse<{ id: string }>>(
    `stories/${storyId}/duplicate`,
    null
  );
  return response;
};
