"use server";

import type { ApiResponse } from "@/types";
import ky from "ky";
import { getApiUrl } from "@/lib/api-url";
import { auth } from "@/auth";
import { getApiError } from "@/utils";

const apiUrl = getApiUrl();

export async function acceptInvitation(inviteToken: string) {
  const session = await auth();
  try {
    const response = await ky.post(
      `${apiUrl}/invitations/${inviteToken}/accept`,
      {
        headers: {
          Authorization: `Bearer ${session?.token}`,
        },
      },
    );

    const data = await response.json<ApiResponse<null>>();
    return data;
  } catch (error) {
    return getApiError(error);
  }
}
