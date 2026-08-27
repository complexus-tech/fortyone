import { auth } from "@/auth";
import { remove } from "@/lib/http";
import type { ApiResponse } from "@/types";
import { executeStoryDeletionRequest } from "./execute-story-deletion-request";

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
