"use server";

import type { ApiResponse, User } from "@/types";
import ky from "ky";
import { auth } from "@/auth";
import { getApiUrl } from "@/lib/api-url";
import { buildAuthHeaders } from "@/lib/http/auth-headers";
import { getCookieHeader } from "@/lib/http/header";
import { getApiError } from "@/utils";

const apiURL = getApiUrl();

export const deleteProfileImageAction = async () => {
  try {
    const session = await auth();
    const cookieHeader = await getCookieHeader();
    const headers = buildAuthHeaders({
      token: session?.token,
      cookieHeader,
    });
    const res = await ky
      .delete(`${apiURL}/users/profile/image`, {
        credentials: "include",
        headers,
      })
      .json<ApiResponse<User>>();
    return res;
  } catch (error) {
    return getApiError(error);
  }
};
