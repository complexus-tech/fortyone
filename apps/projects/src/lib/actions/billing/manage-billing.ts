"use server";

import { auth } from "@/auth";
import { post } from "@/lib/http";
import type { ApiResponse } from "@/types";
import { getApiError } from "@/utils";

export const manageBilling = async (
  returnUrl = "https://www.fortyone.app/login",
) => {
  try {
    const session = await auth();
    const res = await post<{ returnUrl: string }, ApiResponse<{ url: string }>>(
      "subscriptions/portal",
      {
        returnUrl,
      },
      session!,
    );
    return res;
  } catch (error) {
    return getApiError(error);
  }
};
