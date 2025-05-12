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
}: {
  plan: Plan;
  successUrl: string;
  cancelUrl: string;
}) => {
  try {
    const session = await auth();
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
      session!,
    );
    return res;
  } catch (error) {
    return getApiError(error);
  }
};
