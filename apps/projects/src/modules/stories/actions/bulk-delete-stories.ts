"use server";

import { auth } from "@/auth";
import { remove } from "@/lib/http";
import type { ApiResponse } from "@/types";
import { getApiError } from "@/utils";

type Payload = {
  storyIds: string[];
  hardDelete?: boolean;
};

export const bulkDeleteAction = async (
  { storyIds, hardDelete }: Payload,
  workspaceSlug: string,
) => {
  try {
    const session = await auth();
    const ctx = { session: session!, workspaceSlug };
    const stories = await remove<ApiResponse<Payload>>("stories", ctx, {
      json: { storyIds, hardDelete },
    });
    return stories;
  } catch (error) {
    return getApiError(error);
  }
};
