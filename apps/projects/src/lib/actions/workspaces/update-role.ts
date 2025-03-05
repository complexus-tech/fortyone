"use server";

import { put } from "@/lib/http";
import type { ApiResponse, UserRole } from "@/types";
import { getApiError } from "@/utils";

export const updateUserRoleAction = async ({
  userId,
  role,
}: {
  userId: string;
  role: UserRole;
}): Promise<void> => {
  try {
    await put<{ role: UserRole }, ApiResponse<null>>(`members/${userId}/role`, {
      role,
    });
  } catch (error) {
    const res = getApiError(error);
    throw new Error(res.error?.message || "Failed to update user role");
  }
};
