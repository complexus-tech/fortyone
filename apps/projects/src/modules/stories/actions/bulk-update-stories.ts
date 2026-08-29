import { auth } from "@/auth";
import { put } from "@/lib/http";
import type { DetailedStory } from "@/modules/story/types";
import type { ApiResponse } from "@/types";
import { getApiError } from "@/utils";

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
