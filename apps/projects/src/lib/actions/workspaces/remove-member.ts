"use server";

import { remove } from "@/lib/http";
import type { ApiResponse } from "@/types";
import { getApiError } from "@/utils";
import { auth } from "@/auth";

export const removeMemberAction = async (
  memberId: string,
  workspaceSlug: string,
) => {
  try {
    const session = await auth();
    const ctx = { session: session!, workspaceSlug };

    const res = await remove<ApiResponse<null>>(`members/${memberId}`, ctx);
    return res;
  } catch (error) {
    return getApiError(error);
  }
};
