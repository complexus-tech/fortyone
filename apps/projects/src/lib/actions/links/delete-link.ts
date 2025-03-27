"use server";

import { remove } from "@/lib/http";
import type { ApiResponse } from "@/types";
import { getApiError } from "@/utils";

export const deleteLinkAction = async (linkId: string) => {
  try {
    const res = await remove<ApiResponse<void>>(`links/${linkId}`);
    return res;
  } catch (error) {
    return getApiError(error);
  }
};
