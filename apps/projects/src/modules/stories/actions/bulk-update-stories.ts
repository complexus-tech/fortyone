import { auth } from "@/auth";
import { put } from "@/lib/http";
import type { ApiResponse } from "@/types";
import { getApiError } from "@/utils";
import type { DetailedStory } from "../types";

const BULK_UPDATE_TIMEOUT_MS = 60_000;

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
      { timeout: BULK_UPDATE_TIMEOUT_MS, retry: 0 },
    );
    return stories;
  } catch (error) {
    const response = getApiError(error);
    if (error instanceof Error && error.name === "TimeoutError") {
      return {
        data: null,
        error: {
          message:
            "The bulk update timed out. Refresh the stories to check which changes were saved before retrying.",
        },
      };
    }
    return response;
  }
};
