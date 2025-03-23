import { put } from "@/lib/http";
import type { ApiResponse } from "@/types";
import { getApiError } from "@/utils";
import type { UpdateSprint } from "../types";

export const updateSprintAction = async (
  sprintId: string,
  updates: UpdateSprint,
) => {
  try {
    const res = await put<UpdateSprint, ApiResponse<null>>(
      `sprints/${sprintId}`,
      updates,
    );
    return res;
  } catch (error) {
    return getApiError(error);
  }
};
