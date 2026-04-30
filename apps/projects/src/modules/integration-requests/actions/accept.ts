import { post } from "@/lib/http";
import type { ApiResponse } from "@/types";
import { getApiError } from "@/utils";
import { auth } from "@/auth";
import type { IntegrationRequest } from "../types";

export const acceptIntegrationRequestAction = async (
  requestId: string,
  workspaceSlug: string,
) => {
  try {
    const session = await auth();
    const ctx = { session: session!, workspaceSlug };
    return await post<Record<string, never>, ApiResponse<IntegrationRequest>>(
      `integration-requests/${requestId}/accept`,
      {},
      ctx,
    );
  } catch (error) {
    return getApiError(error);
  }
};
