"use server";

import { remove } from "@/lib/http";
import type { ApiResponse } from "@/types";
import { getApiError } from "@/utils";

export const deleteLabelAction = async (labelId: string) => {
  try {
    await remove<ApiResponse<void>>(`labels/${labelId}`);
    return labelId;
  } catch (error) {
    const res = getApiError(error);
    throw new Error(res.error?.message || "Failed to delete label");
  }
};
