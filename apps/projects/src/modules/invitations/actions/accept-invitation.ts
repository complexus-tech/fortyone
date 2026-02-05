"use server";

import type { ApiResponse } from "@/types";
import ky from "ky";
import { buildAuthHeaders } from "@/lib/http/auth-headers";
import { getCookieHeader } from "@/lib/http/header";
import { getApiUrl } from "@/lib/api-url";
import { auth } from "@/auth";
import { getApiError } from "@/utils";

const apiUrl = getApiUrl();

export async function acceptInvitation(inviteToken: string) {
  const session = await auth();
  const cookieHeader = await getCookieHeader();
  try {
    const headers = buildAuthHeaders({
      token: session?.token,
      cookieHeader,
    });
    const response = await ky.post(
      `${apiUrl}/invitations/${inviteToken}/accept`,
      {
        credentials: "include",
        headers,
      },
    );

    const data = await response.json<ApiResponse<null>>();
    return data;
  } catch (error) {
    return getApiError(error);
  }
}
