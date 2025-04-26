"use server";

import { post } from "@/lib/http";
import type { ApiResponse } from "@/types";
import { getApiError } from "@/utils";

export type Plan =
  | "pro_monthly"
  | "pro_yearly"
  | "business_monthly"
  | "business_yearly";
export const checkout = async (plan: Plan) => {
  try {
    const res = await post<
      { priceLookupKey: Plan },
      ApiResponse<{ url: string }>
    >("subscriptions/checkout", {
      priceLookupKey: plan,
    });
    return res;
  } catch (error) {
    return getApiError(error);
  }
};
