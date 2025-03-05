"use server";

import { put } from "@/lib/http";
import { getApiError } from "@/utils";
import type { ObjectiveUpdate } from "../types";

export const updateObjective = async (
  objectiveId: string,
  params: ObjectiveUpdate,
) => {
  try {
    await put<ObjectiveUpdate, null>(`objectives/${objectiveId}`, params);
  } catch (error) {
    const res = getApiError(error);
    throw new Error(res.error?.message || "Failed to update objective");
  }
};
