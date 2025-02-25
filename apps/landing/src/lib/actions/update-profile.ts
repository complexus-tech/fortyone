"use server";
import ky from "ky";
import type { ApiResponse, User } from "@/types";
import { auth } from "@/auth";

const apiUrl = process.env.NEXT_PUBLIC_API_URL;

export type UpdateProfile = {
  fullName?: string;
  username?: string;
};

export async function updateProfile(updates: UpdateProfile) {
  const session = await auth();
  const res = await ky.put(`${apiUrl}/users/profile`, {
    json: updates,
    headers: {
      Authorization: `Bearer ${session?.token}`,
    },
  });
  return res.json<ApiResponse<User>>();
}
