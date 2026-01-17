"use server";

import { auth } from "@/auth";
import { post } from "@/lib/http";
import type { ApiResponse } from "@/types";
import { getApiError } from "@/utils";
import type { Plan } from "./types";

export const checkout = async ({
  plan,
  successUrl,
  cancelUrl,
  workspaceSlug,
}: {
  plan: Plan;
  successUrl: string;
  cancelUrl: string;
  workspaceSlug: string;
}) => {
  try {
    const session = await auth();
    const ctx = { session: session!, workspaceSlug };
    const res = await post<
      { priceLookupKey: Plan; successUrl?: string; cancelUrl?: string },
      ApiResponse<{ url: string }>
    >(
      "subscriptions/checkout",
      {
        priceLookupKey: plan,
        successUrl,
        cancelUrl,
      },
      ctx,
    );
    return res;
  } catch (error) {
    return getApiError(error);
  }
};
