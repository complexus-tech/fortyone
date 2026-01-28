"use server";

import type { ApiResponse, UserRole } from "@/types";
import ky from "ky";
import { getApiUrl } from "@/lib/api-url";
import { getApiError } from "@/utils";

const apiURL = getApiUrl();

type LoginResponse = {
  id: string;
  username: string;
  email: string;
  fullName: string;
  avatarUrl: string;
  lastUsedWorkspaceId: string;
  token: string;
};

export async function authenticateWithToken({
  email,
  token,
}: {
  email: string;
  token: string;
}) {
  try {
    const res = await ky
      .post(`${apiURL}/users/verify/email/confirm`, {
        json: {
          email,
          token,
        },
      })
      .json<ApiResponse<LoginResponse>>();

    const user = res.data!;
    return {
      data: {
        id: user.id,
        name: user.fullName || user.username,
        email: user.email,
        token: user.token,
        workspaces: [],
        image: user.avatarUrl,
        lastUsedWorkspaceId: user.lastUsedWorkspaceId,
        userRole: "guest" as UserRole,
      },
      error: null,
    };
  } catch (error) {
    return getApiError(error);
  }
}
