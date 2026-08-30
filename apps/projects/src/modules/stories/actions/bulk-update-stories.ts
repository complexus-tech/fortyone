import { auth } from "@/auth";
import { put } from "@/lib/http";
import type { ApiResponse } from "@/types";
import { getApiError } from "@/utils";
import type { DetailedStory } from "../types";

type Payload = {
  storyIds: string[];
  updates: Partial<DetailedStory>;
};

export type BulkStoryUpdateResult = {
  totalCount: number;
  succeededCount: number;
  failedCount: number;
  partial: boolean;
  items: {
    storyId: string;
    success: boolean;
    error?: string;
  }[];
};

export const bulkUpdateAction = async (
  updates: Payload,
  workspaceSlug: string,
) => {
  try {
    const session = await auth();
    const ctx = { session: session!, workspaceSlug };
    const stories = await put<Payload, ApiResponse<BulkStoryUpdateResult>>(
      "stories",
      updates,
      ctx,
    );
    return stories;
  } catch (error) {
    return getApiError(error);
  }
};
