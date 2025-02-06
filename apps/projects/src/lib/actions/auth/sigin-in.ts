"use server";
import ky from "ky";
import type { ApiResponse, UserRole } from "@/types";

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

export async function authenticateUser({
  email,
  password,
}: {
  email: string;
  password: string;
}) {
  const res = await ky
    .post(`${apiURL}/users/login`, {
      json: {
        email,
        password,
      },
    })
    .json<ApiResponse<LoginResponse>>();

  if (res.error) {
    throw new Error(res.error.message);
  }
  const user = res.data!;
  return {
    id: user.id,
    name: user.fullName,
    email: user.email,
    token: user.token,
    workspaces: [],
    image: user.avatarUrl,
    lastUsedWorkspaceId: user.lastUsedWorkspaceId,
    userRole: "guest" as UserRole,
  };
}
