"use server";

import type { ApiResponse, User } from "@/types";
import ky from "ky";
import { auth } from "@/auth";
import { getApiUrl } from "@/lib/api-url";
import { buildAuthHeaders } from "@/lib/http/auth-headers";
import { getCookieHeader } from "@/lib/http/header";
import { getApiError } from "@/utils";

const apiURL = getApiUrl();

export const uploadProfileImageAction = async (file: File) => {
  try {
    const formData = new FormData();
    formData.append("image", file);
    const session = await auth();
    const cookieHeader = await getCookieHeader();
    const headers = buildAuthHeaders({
      token: session?.token,
      cookieHeader,
    });
    const res = await ky
      .post(`${apiURL}/users/profile/image`, {
        body: formData,
        credentials: "include",
        headers,
      })
      .json<ApiResponse<User>>();
    return res;
  } catch (error) {
    return getApiError(error);
  }
};
