"use server";

import { remove } from "@/lib/http";
import type { ApiResponse } from "@/types";
import { getApiError } from "@/utils";

export const deleteTeamAction = async (id: string) => {
  try {
    await remove<ApiResponse<void>>(`teams/${id}`);
  } catch (error) {
    const res = getApiError(error);
    throw new Error(res.error?.message || "Failed to delete team");
  }
};
