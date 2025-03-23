import { remove } from "@/lib/http";
import type { ApiResponse } from "@/types";
import { getApiError } from "@/utils";

export const deleteLabelAction = async (labelId: string) => {
  try {
    const res = await remove<ApiResponse<void>>(`labels/${labelId}`);
    return res;
  } catch (error) {
    return getApiError(error);
  }
};
