"use server";

import { auth } from "@/auth";
import { post } from "@/lib/http";
import type { ApiResponse } from "@/types";
import { getApiError } from "@/utils";
import type { Plan } from "./types";

export const changePlan = async (plan: Plan) => {
  try {
    const session = await auth();
    const res = await post<{ newLookupKey: Plan }, ApiResponse<null>>(
      "subscriptions/change-plan",
      {
        newLookupKey: plan,
      },
      session!,
    );
    return res;
  } catch (error) {
    return getApiError(error);
  }
};
