"use server";

import { put } from "@/lib/http";
import type { ApiResponse, UserRole } from "@/types";
import { getApiError } from "@/utils";
import { auth } from "@/auth";

export const updateUserRoleAction = async ({
  userId,
  role,
}: {
  userId: string;
  role: UserRole;
}) => {
  try {
    const session = await auth();
    await put<{ role: UserRole }, ApiResponse<null>>(
      `members/${userId}/role`,
      {
        role,
      },
      session!,
    );
  } catch (error) {
    return getApiError(error);
  }
};
