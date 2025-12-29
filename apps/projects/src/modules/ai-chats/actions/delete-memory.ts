"use server";

import { auth } from "@/auth";
import { remove } from "@/lib/http";
import type { ApiResponse } from "@/types";
import { getApiError } from "@/utils";

export const deleteMemoryAction = async (id: string) => {
  try {
    const session = await auth();
    const result = await remove<ApiResponse<null>>(
      `users/memory/${id}`,
      session!,
    );
    return result;
  } catch (error) {
    return getApiError(error);
  }
};
