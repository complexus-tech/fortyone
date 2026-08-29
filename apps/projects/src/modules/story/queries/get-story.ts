import { ApiError } from "api-client";
import { get } from "@/lib/http";
import type { WorkspaceCtx } from "@/lib/http";
import type { ApiResponse } from "@/types";
import type { DetailedStory } from "../types";

type DetailedStoryResponse = Omit<DetailedStory, "collaboratorIds"> & {
  collaboratorIds?: string[] | null;
};

const normalizeDetailedStory = (
  story: DetailedStoryResponse | null | undefined,
): DetailedStory | null | undefined => {
  if (!story) return story;

  return {
    ...story,
    collaboratorIds: story.collaboratorIds ?? [],
  };
};

export const getStory = async (id: string, ctx: WorkspaceCtx) => {
  try {
    const story = await get<ApiResponse<DetailedStoryResponse>>(
      `stories/${id}`,
      ctx,
    );
    return normalizeDetailedStory(story.data);
  } catch (error) {
    if (error instanceof ApiError && error.status === 404) return null;
    throw error;
  }
};

export const getStoryRef = async (ref: string, ctx: WorkspaceCtx) => {
  try {
    const story = await get<ApiResponse<DetailedStoryResponse>>(
      `story-by-ref/${ref}`,
      ctx,
    );
    return normalizeDetailedStory(story.data);
  } catch (error) {
    if (error instanceof ApiError && error.status === 404) return null;
    throw error;
  }
};
