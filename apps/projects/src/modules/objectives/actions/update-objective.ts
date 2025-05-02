"use server";

import { put } from "@/lib/http";
import type { ApiResponse } from "@/types";
import { getApiError } from "@/utils";
import { auth } from "@/auth";
import type { ObjectiveUpdate } from "../types";

export const updateObjective = async (
  objectiveId: string,
  params: ObjectiveUpdate,
) => {
  try {
    const session = await auth();
    const res = await put<ObjectiveUpdate, ApiResponse<null>>(
      `objectives/${objectiveId}`,
      params,
      session!,
    );
    return res;
  } catch (error) {
    return getApiError(error);
  }
};
