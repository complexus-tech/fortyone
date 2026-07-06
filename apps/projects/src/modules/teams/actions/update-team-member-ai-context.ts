import { put } from "@/lib/http";
import type { ApiResponse } from "@/types";
import { getApiError } from "@/utils";
import { auth } from "@/auth";

type UpdateTeamMemberAIContextPayload = {
  roleTitle: string;
  roleDescription: string;
};

export const updateTeamMemberAIContextAction = async (
  teamId: string,
  memberId: string,
  payload: UpdateTeamMemberAIContextPayload,
  workspaceSlug: string,
) => {
  try {
    const session = await auth();
    const ctx = { session: session!, workspaceSlug };
    return await put<UpdateTeamMemberAIContextPayload, ApiResponse<void>>(
      `teams/${teamId}/members/${memberId}/ai-context`,
      payload,
      ctx,
    );
  } catch (error) {
    return getApiError(error);
  }
};
