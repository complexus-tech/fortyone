"use server";

import { auth } from "@/auth";
import { post } from "@/lib/http";
import type { ApiResponse } from "@/types";
import { getApiError } from "@/utils";

export const manageBilling = async (
  workspaceSlug: string,
  returnUrl = "/login",
) => {
  try {
    const session = await auth();
    const ctx = { session: session!, workspaceSlug };
    const res = await post<{ returnUrl: string }, ApiResponse<{ url: string }>>(
      "subscriptions/portal",
      {
        returnUrl,
      },
      ctx,
    );
    return res;
  } catch (error) {
    return getApiError(error);
  }
};
