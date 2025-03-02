"use server";

import { revalidateTag } from "next/cache";
import { remove } from "@/lib/http";
import { getApiError } from "@/utils";
import { objectiveTags } from "../constants";

export const deleteKeyResult = async (
  keyResultId: string,
  objectiveId: string,
) => {
  try {
    await remove(`key-results/${keyResultId}`);
    revalidateTag(objectiveTags.keyResults(objectiveId));
  } catch (error) {
    const res = getApiError(error);
    throw new Error(res.error?.message || "Failed to delete key result");
  }
};
