"use server";

import { remove } from "@/lib/http";
import type { ApiResponse } from "@/types";
import { getApiError } from "@/utils";
import { auth } from "@/auth";

export const deleteObjectiveStatusAction = async (
  statusId: string,
  workspaceSlug: string,
) => {
  try {
    const session = await auth();
    const ctx = { session: session!, workspaceSlug };
    const response = await remove<ApiResponse<null>>(
      `objective-statuses/${statusId}`,
      ctx,
    );
    return response;
  } catch (error) {
    return getApiError(error);
  }
};
