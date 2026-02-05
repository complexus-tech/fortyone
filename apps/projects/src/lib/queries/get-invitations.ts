"use server";

import type { ApiResponse } from "@/types";
import type { Invitation } from "@/modules/invitations/types";
import ky from "ky";
import { auth } from "@/auth";
import { getApiUrl } from "@/lib/api-url";
import { buildAuthHeaders } from "@/lib/http/auth-headers";
import { getCookieHeader } from "@/lib/http/header";
import { requestError } from "../fetch-error";

const apiUrl = getApiUrl();

export async function getMyInvitations() {
  const session = await auth();
  const cookieHeader = await getCookieHeader();
  if (!session) {
    return {
      data: [],
    } as ApiResponse<Invitation[]>;
  }
  try {
    const headers = buildAuthHeaders({
      token: session?.token,
      cookieHeader,
    });
    const response = await ky.get(`${apiUrl}/users/me/invitations`, {
      credentials: "include",
      headers,
    });
    const data = await response.json<ApiResponse<Invitation[]>>();
    return data;
  } catch (error) {
    const res = await requestError<Invitation[]>(error);
    return {
      data: null,
      error: {
        message: res.error?.message || "Failed to verify invitation",
      },
    };
  }
}
