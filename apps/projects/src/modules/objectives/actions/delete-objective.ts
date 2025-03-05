"use server";

import { remove } from "@/lib/http";
import { getApiError } from "@/utils";

export const deleteObjective = async (objectiveId: string) => {
  try {
    await remove(`objectives/${objectiveId}`);
  } catch (error) {
    const res = getApiError(error);
    throw new Error(res.error?.message || "Failed to delete objective");
  }
};
