"use server";

import { post } from "@/lib/http";
import { getApiError } from "@/utils";
import type { NewObjectiveKeyResult } from "../types";

export const createKeyResult = async (params: NewObjectiveKeyResult) => {
  try {
    await post<NewObjectiveKeyResult, null>("key-results", params);
  } catch (error) {
    const res = getApiError(error);
    throw new Error(res.error?.message || "Failed to create key result");
  }
};
