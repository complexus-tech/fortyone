import { put } from "@/lib/http";
import type { ApiResponse, User } from "@/types";

export type UpdateProfile = {
  fullName?: string;
  username?: string;
};

export const updateProfile = async (updates: UpdateProfile) => {
  const response = await put<UpdateProfile, ApiResponse<User>>(
    "users/profile",
    updates,
    { useWorkspace: false }
  );
  return response.data!;
};
