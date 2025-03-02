"use server";
import { revalidateTag } from "next/cache";
import { put } from "@/lib/http";
import { memberTags } from "@/constants/keys";
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
    revalidateTag(memberTags.lists());
  } catch (error) {
    const res = getApiError(error);
    throw new Error(res.error?.message || "Failed to update user role");
  }
};
