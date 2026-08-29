import { post } from "@/lib/http";
import type { ApiResponse } from "@/types";
import { getApiError } from "@/utils";
import type { NewInvitation } from "../types";

export const inviteMembers = async (
  invites: NewInvitation[],
  workspaceSlug: string,
) => {
  try {
    const response = await post<
      { invitations: NewInvitation[] },
      ApiResponse<null>
    >("invitations", { invitations: invites }, { workspaceSlug });
    return response;
  } catch (error) {
    return getApiError(error);
  }
};
