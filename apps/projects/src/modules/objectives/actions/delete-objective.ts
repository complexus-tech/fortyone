"use server";

import { revalidateTag } from "next/cache";
import { remove } from "@/lib/http";
import { getApiError } from "@/utils";
import { objectiveTags } from "../constants";

export const deleteObjective = async (objectiveId: string) => {
  try {
    await remove(`objectives/${objectiveId}`);
    revalidateTag(objectiveTags.list());
    revalidateTag(objectiveTags.objective(objectiveId));
  } catch (error) {
    const res = getApiError(error);
    throw new Error(res.error?.message || "Failed to delete objective");
  }
};
