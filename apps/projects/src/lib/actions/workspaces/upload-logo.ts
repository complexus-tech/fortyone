"use server";

import { post } from "@/lib/http";
import type { ApiResponse, User } from "@/types";
import { getApiError } from "@/utils";
import { auth } from "@/auth";

export const uploadWorkspaceLogoAction = async (
  file: File,
  workspaceSlug: string,
) => {
  try {
    const formData = new FormData();
    formData.append("image", file);
    const session = await auth();
    const ctx = { session: session!, workspaceSlug };
    const res = await post<FormData, ApiResponse<User>>("logo", formData, ctx);
    return res;
  } catch (error) {
    return getApiError(error);
  }
};
