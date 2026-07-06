import { post } from "@/lib/http";
import type { ApiResponse } from "@/types";
import { getApiError } from "@/utils";
import { auth } from "@/auth";
import type { BulkIntegrationRequestResult } from "../types";

export const acceptAllIntegrationRequestsAction = async (
  teamId: string,
  workspaceSlug: string,
) => {
  try {
    const session = await auth();
    const ctx = { session: session!, workspaceSlug };
    return await post<
      Record<string, never>,
      ApiResponse<BulkIntegrationRequestResult>
    >(`teams/${teamId}/integration-requests/accept-all`, {}, ctx);
  } catch (error) {
    return getApiError(error);
  }
};
