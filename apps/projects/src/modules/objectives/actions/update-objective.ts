"use server";

import { revalidateTag } from "next/cache";
import { put } from "@/lib/http";
import { getApiError } from "@/utils";
import type { ObjectiveUpdate } from "../types";
import { objectiveTags } from "../constants";

export const updateObjective = async (
  objectiveId: string,
  params: ObjectiveUpdate,
) => {
  try {
    await put<ObjectiveUpdate, null>(`objectives/${objectiveId}`, params);
    revalidateTag(objectiveTags.list());
    revalidateTag(objectiveTags.objective(objectiveId));
  } catch (error) {
    const res = getApiError(error);
    throw new Error(res.error?.message || "Failed to update objective");
  }
};
