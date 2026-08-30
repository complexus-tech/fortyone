import { auth } from "@/auth";
import { remove } from "@/lib/http";
import { executeStoryDeletionRequest } from "@/shared/story/deletion";
import type { ApiResponse } from "@/types";

export const deleteStoryAction = async (
  storyId: string,
  workspaceSlug: string,
) => {
  const session = await auth();
  const ctx = { session: session!, workspaceSlug };

  return executeStoryDeletionRequest({
    request: () => remove<ApiResponse<null>>(`stories/${storyId}`, ctx),
    retryUncertain: true,
  });
};
