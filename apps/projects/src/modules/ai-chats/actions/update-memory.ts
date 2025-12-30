"use server";

import { auth } from "@/auth";
import { put } from "@/lib/http";
import type { ApiResponse } from "@/types";
import { getApiError } from "@/utils";
import type { UpdateMemoryPayload } from "../types";

export const updateMemoryAction = async (
  id: string,
  payload: UpdateMemoryPayload,
) => {
  try {
    const session = await auth();
    const result = await put<UpdateMemoryPayload, ApiResponse<null>>(
      `users/memory/${id}`,
      payload,
      session!,
    );
    return result;
  } catch (error) {
    return getApiError(error);
  }
};
