import type { WorkspaceCtx } from "@/lib/http";
import type { ApiResponse } from "@/types";
import { get } from "@/lib/http";
import type { StoryChecklistItem } from "../types/checklist";

export const getStoryChecklist = async (
  storyId: string,
  ctx: WorkspaceCtx,
  signal?: AbortSignal,
): Promise<StoryChecklistItem[]> => {
  const response = await get<ApiResponse<StoryChecklistItem[]>>(
    `stories/${storyId}/checklist`,
    ctx,
    { signal },
  );
  if (response.error) throw new Error(response.error.message);
  if (!Array.isArray(response.data)) {
    throw new Error("Unable to load the checklist. Please try again.");
  }
  return response.data;
};
