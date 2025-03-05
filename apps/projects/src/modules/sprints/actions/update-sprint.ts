"use server";

import { put } from "@/lib/http";
import { getApiError } from "@/utils";
import type { UpdateSprint } from "../types";

export const updateSprintAction = async (
  sprintId: string,
  updates: UpdateSprint,
) => {
  try {
    await put(`sprints/${sprintId}`, updates);
  } catch (error) {
    const res = getApiError(error);
    throw new Error(res.error?.message || "Failed to update sprint");
  }
};
