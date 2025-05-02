"use server";

import { remove } from "@/lib/http";
import type { ApiResponse } from "@/types";
import { getApiError } from "@/utils";
import { auth } from "@/auth";

export const deleteSprintAction = async (sprintId: string) => {
  try {
    const session = await auth();
    const res = await remove<ApiResponse<null>>(
      `sprints/${sprintId}`,
      session!,
    );
    return res;
  } catch (error) {
    return getApiError(error);
  }
};
