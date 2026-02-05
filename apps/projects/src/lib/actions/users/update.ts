"use server";

import type { ApiResponse, User } from "@/types";
import ky from "ky";
import { auth } from "@/auth";
import { getApiUrl } from "@/lib/api-url";
import { buildAuthHeaders } from "@/lib/http/auth-headers";
import { getCookieHeader } from "@/lib/http/header";
import { getApiError } from "@/utils";

export type UpdateProfile = {
  fullName?: string;
  username?: string;
  hasSeenWalkthrough?: boolean;
  timezone?: string;
};

const apiURL = getApiUrl();

export async function updateProfile(updates: UpdateProfile) {
  const session = await auth();
  const cookieHeader = await getCookieHeader();
  try {
    const headers = buildAuthHeaders({
      token: session?.token,
      cookieHeader,
    });
    const res = await ky.put(`${apiURL}/users/profile`, {
      json: updates,
      credentials: "include",
      headers,
    });
    return res.json<ApiResponse<User>>();
  } catch (error) {
    const res = getApiError(error);
    throw new Error(res.error?.message || "Failed to update profile");
  }
}
