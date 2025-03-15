"use server";

import { put } from "@/lib/http";
import type { ApiResponse } from "@/types";
import { getApiError } from "@/utils";
import type { ObjectiveUpdate } from "../types";

export const updateObjective = async (
  objectiveId: string,
  params: ObjectiveUpdate,
) => {
  try {
    const res = await put<ObjectiveUpdate, ApiResponse<null>>(
      `objectives/${objectiveId}`,
      params,
    );
    return res;
  } catch (error) {
    return getApiError(error);
  }
};
