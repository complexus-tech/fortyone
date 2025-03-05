"use server";

import { remove } from "@/lib/http";
import { getApiError } from "@/utils";

export const deleteStateAction = async (stateId: string) => {
  try {
    await remove(`states/${stateId}`);
    return stateId;
  } catch (error) {
    const res = getApiError(error);
    throw new Error(res.error?.message || "Failed to delete state");
  }
};
