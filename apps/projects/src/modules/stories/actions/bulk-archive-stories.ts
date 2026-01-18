"use server";

import { post } from "@/lib/http";
import type { ApiResponse } from "@/types";
import { getApiError } from "@/utils";
import { auth } from "@/auth";

type Payload = {
  storyIds: string[];
};

export const bulkArchiveAction = async (
  storyIds: string[],
  workspaceSlug: string,
) => {
  try {
    const session = await auth();
    const ctx = { session: session!, workspaceSlug };
    const stories = await post<Payload, ApiResponse<Payload>>(
      "stories/archive",
      {
        storyIds,
      },
      ctx,
    );

    return stories;
  } catch (error) {
    return getApiError(error);
  }
};
