"use server";

import { remove } from "@/lib/http";
import { getApiError } from "@/utils";

export const deleteKeyResult = async (keyResultId: string) => {
  try {
    await remove(`key-results/${keyResultId}`);
  } catch (error) {
    const res = getApiError(error);
    throw new Error(res.error?.message || "Failed to delete key result");
  }
};
