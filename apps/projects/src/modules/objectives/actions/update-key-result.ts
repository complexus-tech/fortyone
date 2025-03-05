"use server";

import { put } from "@/lib/http";
import { getApiError } from "@/utils";
import type { KeyResultUpdate } from "../types";

export const updateKeyResult = async (
  keyResultId: string,
  objectiveId: string,
  params: KeyResultUpdate,
) => {
  try {
    await put<KeyResultUpdate, null>(`key-results/${keyResultId}`, params);
  } catch (error) {
    const res = getApiError(error);
    throw new Error(res.error?.message || "Failed to update key result");
  }
};
