import { auth } from "@/auth";
import { put } from "@/lib/http";
import type { ApiResponse } from "@/types";
import { getApiError } from "@/utils";
import type { UpdateMemoryPayload } from "../types";

export const updateMemoryAction = async (
  id: string,
  payload: UpdateMemoryPayload,
  workspaceSlug: string,
) => {
  try {
    const session = await auth();
    if (!session) {
      return {
        data: null,
        error: { message: "Authentication required" },
      } satisfies ApiResponse<null>;
    }

    const ctx = { session, workspaceSlug };
    const result = await put<UpdateMemoryPayload, ApiResponse<null>>(
      `users/memory/${id}`,
      payload,
      ctx,
    );
    return result;
  } catch (error) {
    return getApiError(error);
  }
};
