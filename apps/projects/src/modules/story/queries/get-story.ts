import { get } from "@/lib/http";
import type { WorkspaceCtx } from "@/lib/http";
import type { ApiResponse } from "@/types";
import type { DetailedStory } from "../types";

export const getStory = async (id: string, ctx: WorkspaceCtx) => {
  try {
    const story = await get<ApiResponse<DetailedStory>>(`stories/${id}`, ctx);
    return story.data;
  } catch (error) {
    return null;
  }
};

export const getStoryRef = async (ref: string, ctx: WorkspaceCtx) => {
  try {
    const story = await get<ApiResponse<DetailedStory>>(
      `story-by-ref/${ref}`,
      ctx,
    );
    return story.data;
  } catch (error) {
    return null;
  }
};
