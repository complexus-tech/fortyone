"use server";

import type { ApiResponse } from "@/types";
import ky from "ky";
import { auth } from "@/auth";
import { getApiUrl } from "@/lib/api-url";
import { buildAuthHeaders } from "@/lib/http/auth-headers";
import { getCookieHeader } from "@/lib/http/header";
import { requestError } from "../fetch-error";

const apiUrl = getApiUrl();

export async function acceptInvitation(token: string) {
  const session = await auth();
  const cookieHeader = await getCookieHeader();

  try {
    const headers = buildAuthHeaders({
      token: session?.token,
      cookieHeader,
    });
    const response = await ky.post(`${apiUrl}/invitations/${token}/accept`, {
      credentials: "include",
      headers,
    });

    const data = await response.json<ApiResponse<null>>();
    return data;
  } catch (error) {
    return requestError<null>(error);
  }
}
