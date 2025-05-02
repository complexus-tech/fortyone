"use server";

import { auth } from "@/auth";
import { remove } from "@/lib/http";
import type { ApiResponse } from "@/types";
import { getApiError } from "@/utils";

export const deleteObjective = async (objectiveId: string) => {
  try {
    const session = await auth();
    const res = await remove<ApiResponse<null>>(
      `objectives/${objectiveId}`,
      session!,
    );
    return res;
  } catch (error) {
    return getApiError(error);
  }
};
