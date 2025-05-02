"use server";

import { remove } from "@/lib/http";
import type { ApiResponse } from "@/types";
import { getApiError } from "@/utils";
import { auth } from "@/auth";

export const revokeInvitation = async (invitationId: string) => {
  try {
    const session = await auth();
    const response = await remove<ApiResponse<null>>(
      `invitations/${invitationId}`,
      session!,
    );
    return response;
  } catch (error) {
    return getApiError(error);
  }
};
