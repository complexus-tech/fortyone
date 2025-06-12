"use server";

import { remove } from "@/lib/http";
import type { ApiResponse, User } from "@/types";
import { getApiError } from "@/utils";
import { auth } from "@/auth";

export const deleteWorkspaceLogoAction = async () => {
  try {
    const session = await auth();
    const res = await remove<ApiResponse<User>>("logo", session!);
    return res;
  } catch (error) {
    return getApiError(error);
  }
};
