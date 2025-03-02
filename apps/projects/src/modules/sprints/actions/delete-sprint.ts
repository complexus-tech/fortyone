"use server";

import { revalidateTag } from "next/cache";
import { remove } from "@/lib/http";
import { sprintTags } from "@/constants/keys";
import { getApiError } from "@/utils";

export const deleteSprintAction = async (sprintId: string) => {
  try {
    await remove(`sprints/${sprintId}`);
    revalidateTag(sprintTags.lists());
  } catch (error) {
    const res = getApiError(error);
    throw new Error(res.error?.message || "Failed to delete sprint");
  }
};
