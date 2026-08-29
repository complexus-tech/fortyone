import { remove } from "@/lib/http";
import type { ApiResponse } from "@/types";
import { getApiError } from "@/utils";

export const revokeInvitation = async (
  invitationId: string,
  workspaceSlug: string,
) => {
  try {
    const response = await remove<ApiResponse<null>>(
      `invitations/${invitationId}`,
      { workspaceSlug },
    );
    return response;
  } catch (error) {
    return getApiError(error);
  }
};
