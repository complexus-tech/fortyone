"use server";
import type { ApiResponse, User } from "@/types";
import { put } from "@/lib/http";

export type UpdateProfile = {
  fullName: string;
  username: string;
};

export async function updateProfile(updates: UpdateProfile) {
  const res = await put<UpdateProfile, ApiResponse<User>>("profile", updates);

  return res.data;
}
