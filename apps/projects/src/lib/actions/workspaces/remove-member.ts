"use server";

import { remove } from "@/lib/http";
import type { ApiResponse } from "@/types";
import { getApiError } from "@/utils";
import { auth } from "@/auth";

export const removeMemberAction = async (memberId: string) => {
  try {
    const session = await auth();
    await remove<ApiResponse<void>>(`/members/${memberId}`, session!);
  } catch (error) {
    const apiError = getApiError(error);
    throw new Error(apiError.error?.message || "Failed to remove member");
  }
};
