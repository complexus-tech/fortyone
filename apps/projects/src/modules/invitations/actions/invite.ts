"use server";

import { post } from "@/lib/http";
import type { ApiResponse } from "@/types";
import { getApiError } from "@/utils";
import { auth } from "@/auth";
import type { NewInvitation } from "../types";

export const inviteMembers = async (
  invites: NewInvitation[],
  workspaceSlug: string,
) => {
  try {
    const session = await auth();
    const ctx = { session: session!, workspaceSlug };
    const response = await post<{ invitations: NewInvitation[] }, ApiResponse<null>>(
      "invitations",
      { invitations: invites },
      ctx,
    );
    return response;
  } catch (error) {
    return getApiError(error);
  }
};
