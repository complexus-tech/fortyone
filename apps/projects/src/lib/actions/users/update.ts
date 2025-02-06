"use server";
import { revalidateTag } from "next/cache";
import type { ApiResponse, User } from "@/types";
import { put } from "@/lib/http";
import { userTags } from "@/constants/keys";

export type UpdateProfile = {
  fullName: string;
  username: string;
};

export async function updateProfile(updates: UpdateProfile) {
  const res = await put<UpdateProfile, ApiResponse<User>>("profile", updates);
  revalidateTag(userTags.profile());
  return res.data;
}
