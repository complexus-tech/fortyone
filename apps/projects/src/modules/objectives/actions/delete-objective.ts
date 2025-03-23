import { remove } from "@/lib/http";
import type { ApiResponse } from "@/types";
import { getApiError } from "@/utils";

export const deleteObjective = async (objectiveId: string) => {
  try {
    const res = await remove<ApiResponse<null>>(`objectives/${objectiveId}`);
    return res;
  } catch (error) {
    return getApiError(error);
  }
};
