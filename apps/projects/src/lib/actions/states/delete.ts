"use server";

import { revalidateTag } from "next/cache";
import { statusTags } from "@/constants/keys";
import { remove } from "@/lib/http";
import { getApiError } from "@/utils";

export const deleteStateAction = async (stateId: string) => {
  try {
    await remove(`states/${stateId}`);
    revalidateTag(statusTags.lists());
    return stateId;
  } catch (error) {
    const res = getApiError(error);
    throw new Error(res.error?.message || "Failed to delete state");
  }
};
