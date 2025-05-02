"use server";

import { post } from "@/lib/http";
import type { ApiResponse } from "@/types";
import { getApiError } from "@/utils";
import { auth } from "@/auth";
import type { NewObjectiveKeyResult } from "../types";

export const createKeyResult = async (params: NewObjectiveKeyResult) => {
  try {
    const session = await auth();
    const res = await post<NewObjectiveKeyResult, ApiResponse<null>>(
      "key-results",
      params,
      session!,
    );
    return res;
  } catch (error) {
    return getApiError(error);
  }
};
