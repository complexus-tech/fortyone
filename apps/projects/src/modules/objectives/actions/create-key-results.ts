import { post } from "@/lib/http";
import type { ApiResponse } from "@/types";
import { getApiError } from "@/utils";
import { auth } from "@/auth";
import type { KeyResult, NewKeyResult } from "../types";

export const createKeyResults = async (
  objectiveId: string,
  keyResults: NewKeyResult[],
  workspaceSlug: string,
) => {
  try {
    const session = await auth();
    const ctx = { session: session!, workspaceSlug };
    return await post<{ keyResults: NewKeyResult[] }, ApiResponse<KeyResult[]>>(
      `objectives/${objectiveId}/key-results`,
      { keyResults },
      ctx,
    );
  } catch (error) {
    return getApiError(error);
  }
};
