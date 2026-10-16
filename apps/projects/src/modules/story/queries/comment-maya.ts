import type { WorkspaceCtx } from "@/lib/http";
import type { ApiResponse } from "@/types/api-response";
import { get, post } from "@/lib/http";
import type { CommentMayaAction, StoryMayaRuns } from "../types/comment-maya";

export const getStoryMayaRuns = async (storyId: string, ctx: WorkspaceCtx) => {
  const response = await get<ApiResponse<StoryMayaRuns>>(
    `stories/${storyId}/maya-runs`,
    ctx,
  );
  if (response.error) throw new Error(response.error.message);
  if (!response.data) throw new Error("Maya requests are unavailable.");
  return response.data;
};

export const updateCommentMayaRun = async (
  storyId: string,
  runId: string,
  action: CommentMayaAction,
  version: number,
  ctx: WorkspaceCtx,
) => {
  const response = await post<{ version: number }, ApiResponse<unknown>>(
    `stories/${storyId}/maya-runs/${runId}/${action}`,
    { version },
    ctx,
  );
  if (response.error) throw new Error(response.error.message);
};
