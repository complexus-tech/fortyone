"use server";

import type { ApiResponse, User } from "@/types";
import ky from "ky";
import { getApiUrl } from "@/lib/api-url";
import { auth } from "@/auth";
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
  try {
    const res = await ky.put(`${apiURL}/users/profile`, {
      json: updates,
      headers: {
        Authorization: `Bearer ${session?.token}`,
      },
    });
    return res.json<ApiResponse<User>>();
  } catch (error) {
    const res = getApiError(error);
    throw new Error(res.error?.message || "Failed to update profile");
  }
}
