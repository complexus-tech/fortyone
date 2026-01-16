"use server";

import ky from "ky";
import type { User } from "next-auth";
import type { ApiResponse, UserRole } from "@/types";
import { requestError } from "../fetch-error";

const apiURL = process.env.NEXT_PUBLIC_API_URL;

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
    const data = await requestError<User>(error);
    return data;
  }
}

export async function authenticateGoogleUser({ idToken }: { idToken: string }) {
  const res = await ky
    .post(`${apiURL}/users/google/verify`, {
      json: {
        token: idToken,
      },
    })
    .json<ApiResponse<LoginResponse>>();

  if (res.error) {
    throw new Error(res.error.message);
  }

  const user = res.data!;
  return {
    id: user.id,
    name: user.fullName || user.username,
    email: user.email,
    token: user.token,
    workspaces: [],
    image: user.avatarUrl,
    lastUsedWorkspaceId: user.lastUsedWorkspaceId,
    userRole: "guest" as UserRole,
  };
}
