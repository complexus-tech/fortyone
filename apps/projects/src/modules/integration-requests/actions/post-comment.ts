import { post } from "@/lib/http";
import type { ApiResponse } from "@/types";
import { getApiError } from "@/utils";
import { auth } from "@/auth";
import type { IntegrationRequestComment } from "../types";

export const postIntegrationRequestCommentAction = async (
  requestId: string,
  body: string,
  idempotencyKey: string,
  workspaceSlug: string,
) => {
  try {
    const session = await auth();
    return await post<
      { body: string; idempotencyKey: string },
      ApiResponse<IntegrationRequestComment>
    >(
      `integration-requests/${requestId}/comments`,
      { body, idempotencyKey },
      { session: session!, workspaceSlug },
    );
  } catch (error) {
    return getApiError(error);
  }
};
