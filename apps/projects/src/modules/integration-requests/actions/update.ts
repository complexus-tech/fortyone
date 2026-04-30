import { put } from "@/lib/http";
import type { ApiResponse } from "@/types";
import { getApiError } from "@/utils";
import { auth } from "@/auth";
import type {
  IntegrationRequest,
  UpdateIntegrationRequestInput,
} from "../types";

export const updateIntegrationRequestAction = async (
  requestId: string,
  payload: UpdateIntegrationRequestInput,
  workspaceSlug: string,
) => {
  try {
    const session = await auth();
    const ctx = { session: session!, workspaceSlug };
    return await put<
      UpdateIntegrationRequestInput,
      ApiResponse<IntegrationRequest>
    >(`integration-requests/${requestId}`, payload, ctx);
  } catch (error) {
    return getApiError(error);
  }
};
