"use server";

import { auth } from "@/auth";
import { post } from "@/lib/http";
import type { ApiResponse } from "@/types";
import { getApiError } from "@/utils";
import type { CreateMemoryPayload, Memory } from "../types";

export const createMemoryAction = async (
  payload: CreateMemoryPayload,
  workspaceSlug: string,
) => {
  try {
    const session = await auth();
    const ctx = { session: session!, workspaceSlug };
    const result = await post<CreateMemoryPayload, ApiResponse<Memory>>(
      "users/memory",
      payload,
      ctx,
    );
    return result;
  } catch (error) {
    return getApiError(error);
  }
};
