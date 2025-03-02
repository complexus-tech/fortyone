"use server";

import { revalidateTag } from "next/cache";
import { put } from "@/lib/http";
import { sprintTags } from "@/constants/keys";
import { getApiError } from "@/utils";
import type { UpdateSprint } from "../types";

export const updateSprintAction = async (
  sprintId: string,
  updates: UpdateSprint,
) => {
  try {
    await put(`sprints/${sprintId}`, updates);
    revalidateTag(sprintTags.lists());
  } catch (error) {
    const res = getApiError(error);
    throw new Error(res.error?.message || "Failed to update sprint");
  }
};
