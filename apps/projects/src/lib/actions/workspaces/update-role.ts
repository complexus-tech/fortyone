import { put } from "@/lib/http";
import type { ApiResponse, UserRole } from "@/types";
import { getApiError } from "@/utils";

export const updateUserRoleAction = async ({
  userId,
  role,
}: {
  userId: string;
  role: UserRole;
}) => {
  try {
    await put<{ role: UserRole }, ApiResponse<null>>(`members/${userId}/role`, {
      role,
    });
  } catch (error) {
    return getApiError(error);
  }
};
