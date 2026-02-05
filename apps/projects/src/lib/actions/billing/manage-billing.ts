"use server";

import { auth } from "@/auth";
import { post } from "@/lib/http";
import { getCookieHeader } from "@/lib/http/header";
import type { ApiResponse } from "@/types";
import { getApiError } from "@/utils";

export const manageBilling = async (workspaceSlug: string, returnUrl = "/") => {
  try {
    const session = await auth();
    const cookieHeader = await getCookieHeader();
    const ctx = { session: session!, workspaceSlug, cookieHeader };
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
