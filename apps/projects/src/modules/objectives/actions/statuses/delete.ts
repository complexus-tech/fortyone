"use server";

import { remove } from "@/lib/http";
import type { ApiResponse } from "@/types";
import { getApiError } from "@/utils";

export const deleteObjectiveStatusAction = async (statusId: string) => {
  try {
    await remove<ApiResponse<void>>(`objective-statuses/${statusId}`);
  } catch (error) {
    const res = getApiError(error);
    throw new Error(res.error?.message || "Failed to delete objective status");
  }
};
