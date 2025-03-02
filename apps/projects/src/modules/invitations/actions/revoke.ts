"use server";

import { remove } from "@/lib/http";
import { getApiError } from "@/utils";

export const revokeInvitation = async (invitationId: string) => {
  try {
    const response = await remove(`invitations/${invitationId}`);
    return response;
  } catch (error) {
    const res = getApiError(error);
    return res;
  }
};
