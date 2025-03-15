"use server";

import { remove } from "@/lib/http";
import type { ApiResponse } from "@/types";
import { getApiError } from "@/utils";

export const deleteSprintAction = async (sprintId: string) => {
  try {
    const res = await remove<ApiResponse<null>>(`sprints/${sprintId}`);
    return res;
  } catch (error) {
    return getApiError(error);
  }
};
