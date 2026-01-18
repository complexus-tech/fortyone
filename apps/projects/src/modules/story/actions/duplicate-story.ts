"use server";

import { post } from "@/lib/http";
import type { ApiResponse } from "@/types";
import { getApiError } from "@/utils";
import { auth } from "@/auth";
import type { DetailedStory } from "../types";

export const duplicateStoryAction = async (
  storyId: string,
  workspaceSlug: string,
) => {
  try {
    const session = await auth();
    const ctx = { session: session!, workspaceSlug };
    const story = await post<Record<string, never>, ApiResponse<DetailedStory>>(
      `stories/${storyId}/duplicate`,
      {},
      ctx,
    );
    return story;
  } catch (error) {
    return getApiError(error);
  }
};
