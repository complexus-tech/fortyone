"use server";

import { revalidateTag } from "next/cache";
import { post } from "@/lib/http";
import { getApiError } from "@/utils";
import type { NewObjectiveKeyResult } from "../types";
import { objectiveTags } from "../constants";

export const createKeyResult = async (params: NewObjectiveKeyResult) => {
  try {
    await post<NewObjectiveKeyResult, null>("key-results", params);
    revalidateTag(objectiveTags.keyResults(params.objectiveId));
  } catch (error) {
    const res = getApiError(error);
    throw new Error(res.error?.message || "Failed to create key result");
  }
};
