"use server";

import type { ApiResponse } from "@/types";
import ky from "ky";
import { auth } from "@/auth";
import { getApiUrl } from "@/lib/api-url";
import { buildAuthHeaders } from "@/lib/http/auth-headers";
import { getCookieHeader } from "@/lib/http/header";
import { requestError } from "../fetch-error";

const apiUrl = getApiUrl();

export async function checkWorkspaceAvailability(slug: string) {
  const session = await auth();
  const cookieHeader = await getCookieHeader();
  try {
    const headers = buildAuthHeaders({
      token: session?.token,
      cookieHeader,
    });
    const availability = await ky
      .get(`${apiUrl}/workspaces/check-availability?slug=${slug}`, {
        credentials: "include",
        headers,
      })
      .json<ApiResponse<{ available: boolean; slug: string }>>();

    return availability;
  } catch (error) {
    const data = await requestError<{ available: boolean; slug: string }>(
      error,
    );
    return data;
  }
}
