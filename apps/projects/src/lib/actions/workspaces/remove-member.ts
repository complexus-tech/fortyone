"use server";

import { revalidateTag } from "next/cache";
import { memberTags } from "@/constants/keys";
import { remove } from "@/lib/http";
import type { ApiResponse } from "@/types";
import { getApiError } from "@/utils";

export const removeMemberAction = async (memberId: string) => {
  try {
    await remove<ApiResponse<void>>(`/members/${memberId}`);
    revalidateTag(memberTags.all);
  } catch (error) {
    const apiError = getApiError(error);
    throw new Error(apiError.error?.message || "Failed to remove member");
  }
};
