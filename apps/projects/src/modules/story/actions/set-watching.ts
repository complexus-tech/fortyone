import { put } from "@/lib/http";
import type { ApiResponse } from "@/types";
import { getApiError } from "@/utils";
import { auth } from "@/auth";

export const setStoryWatchingAction = async (
  storyId: string,
  watching: boolean,
  workspaceSlug: string,
) => {
  try {
    const session = await auth();
    return await put<{ watching: boolean }, ApiResponse<null>>(
      `stories/${storyId}/watch`,
      { watching },
      { session: session!, workspaceSlug },
    );
  } catch (error) {
    return getApiError(error);
  }
};
