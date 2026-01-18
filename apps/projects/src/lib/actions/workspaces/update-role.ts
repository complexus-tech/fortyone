"use server";

import { put } from "@/lib/http";
import type { ApiResponse, UserRole } from "@/types";
import { getApiError } from "@/utils";
import { auth } from "@/auth";

export const updateUserRoleAction = async ({
  userId,
  role,
  workspaceSlug,
}: {
  userId: string;
  role: UserRole;
  workspaceSlug: string;
}) => {
  try {
    const session = await auth();
    const ctx = { session: session!, workspaceSlug };
    await put<{ role: UserRole }, ApiResponse<null>>(
      `members/${userId}/role`,
      {
        role,
      },
      ctx,
    );
  } catch (error) {
    return getApiError(error);
  }
};
