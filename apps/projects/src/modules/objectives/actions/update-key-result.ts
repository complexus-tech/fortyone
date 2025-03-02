"use server";

import { revalidateTag } from "next/cache";
import { put } from "@/lib/http";
import { getApiError } from "@/utils";
import type { KeyResultUpdate } from "../types";
import { objectiveTags } from "../constants";

export const updateKeyResult = async (
  keyResultId: string,
  objectiveId: string,
  params: KeyResultUpdate,
) => {
  try {
    await put<KeyResultUpdate, null>(`key-results/${keyResultId}`, params);
    revalidateTag(objectiveTags.keyResults(objectiveId));
  } catch (error) {
    const res = getApiError(error);
    throw new Error(res.error?.message || "Failed to update key result");
  }
};
