import { get } from "@/lib/http";
import type { WorkspaceCtx } from "@/lib/http";
import type { ApiResponse } from "@/types";
import type { StoryAttachment } from "../types";

export const getStoryAttachments = async (id: string, ctx: WorkspaceCtx) => {
  const story = await get<ApiResponse<StoryAttachment[]>>(
    `stories/${id}/attachments`,
    ctx,
  );
  return story.data!;
};
