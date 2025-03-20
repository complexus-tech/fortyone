"use server";

import ky from "ky";
import { auth, refreshWorkspaces } from "@/auth";
import type { ApiResponse } from "@/types";
import { getApiError } from "@/utils";

const apiUrl = process.env.NEXT_PUBLIC_API_URL;

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
    await refreshWorkspaces();

    const data = await response.json<ApiResponse<null>>();
    return data;
  } catch (error) {
    return getApiError(error);
  }
}
