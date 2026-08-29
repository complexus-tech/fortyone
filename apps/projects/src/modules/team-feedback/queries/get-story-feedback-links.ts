import { get } from "@/lib/http";
import type { WorkspaceCtx } from "@/lib/http";
import type { ApiResponse } from "@/types";
import type { StoryFeedbackLink } from "../types";

export const getStoryFeedbackLinks = async (
  storyId: string,
  ctx: WorkspaceCtx,
) => {
  const response = await get<ApiResponse<StoryFeedbackLink[]>>(
    `stories/${storyId}/feedback-links`,
    ctx,
  );

  return response.data ?? [];
};
