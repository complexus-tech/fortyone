"use server";

import { remove } from "@/lib/http";
import type { ApiResponse } from "@/types";
import { getApiError } from "@/utils";

export const deleteStateAction = async (stateId: string) => {
  try {
    const res = await remove<ApiResponse<void>>(`states/${stateId}`);
    return res;
  } catch (error) {
    return getApiError(error);
  }
};
