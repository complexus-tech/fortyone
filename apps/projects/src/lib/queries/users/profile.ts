"use server";
import ky from "ky";
import type { ApiResponse, User } from "@/types";
import { auth } from "@/auth";

const apiURL = process.env.NEXT_PUBLIC_API_URL;

export async function getProfile() {
  const session = await auth();
  const res = await ky.get(`${apiURL}/users/profile`, {
    headers: {
      Authorization: `Bearer ${session?.token}`,
    },
  });
  const data = await res.json<ApiResponse<User>>();
  return data.data!;
}
