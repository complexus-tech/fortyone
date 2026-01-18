"use server";

import { put } from "@/lib/http";
import type { ApiResponse } from "@/types";
import { getApiError } from "@/utils";
import { auth } from "@/auth";
import type { KeyResultUpdate } from "../types";

export const updateKeyResult = async (
  keyResultId: string,
  params: KeyResultUpdate,
  workspaceSlug: string,
) => {
  try {
    const session = await auth();
    const ctx = { session: session!, workspaceSlug };
    const res = await put<KeyResultUpdate, ApiResponse<null>>(
      `key-results/${keyResultId}`,
      params,
      ctx,
    );
    return res;
  } catch (error) {
    return getApiError(error);
  }
};
