"use server";

import { auth } from "@/auth";
import { post } from "@/lib/http";
import type { ApiResponse } from "@/types";
import { getApiError } from "@/utils";
import type { CreateMemoryPayload, Memory } from "../types";

export const createMemoryAction = async (payload: CreateMemoryPayload) => {
  try {
    const session = await auth();
    const result = await post<CreateMemoryPayload, ApiResponse<Memory>>(
      "users/memory",
      payload,
      session!,
    );
    return result;
  } catch (error) {
    return getApiError(error);
  }
};
