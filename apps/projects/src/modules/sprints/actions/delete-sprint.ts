"use server";

import { remove } from "@/lib/http";
import { getApiError } from "@/utils";

export const deleteSprintAction = async (sprintId: string) => {
  try {
    await remove(`sprints/${sprintId}`);
  } catch (error) {
    const res = getApiError(error);
    throw new Error(res.error?.message || "Failed to delete sprint");
  }
};
