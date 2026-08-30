import { auth } from "@/auth";
import { remove } from "@/lib/http";
import { executeStoryDeletionRequest } from "@/shared/story/deletion";
import type { ApiResponse } from "@/types";

type Payload = {
  storyIds: string[];
  hardDelete?: boolean;
};

export type BulkDeleteResult = {
  deletedCount: number;
  storyIds: string[];
};

export const bulkDeleteAction = async (
  { storyIds, hardDelete }: Payload,
  workspaceSlug: string,
) => {
  const session = await auth();
  const ctx = { session: session!, workspaceSlug };

  return executeStoryDeletionRequest({
    request: () =>
      remove<ApiResponse<BulkDeleteResult>>("stories", ctx, {
        json: { storyIds, hardDelete },
      }),
    // Permanent deletion is not retried: a committed first request would make
    // the exact retry return not-found and conceal the original outcome.
    retryUncertain: hardDelete !== true,
  });
};
