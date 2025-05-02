"use server";

import { put } from "@/lib/http";
import type { ApiResponse } from "@/types";
import { getApiError } from "@/utils";
import { auth } from "@/auth";
import type { UpdateSprint } from "../types";

export const updateSprintAction = async (
  sprintId: string,
  updates: UpdateSprint,
) => {
  try {
    const session = await auth();
    const res = await put<UpdateSprint, ApiResponse<null>>(
      `sprints/${sprintId}`,
      updates,
      session!,
    );
    return res;
  } catch (error) {
    return getApiError(error);
  }
};
