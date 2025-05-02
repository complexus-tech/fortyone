"use server";

import { remove } from "@/lib/http";
import type { ApiResponse } from "@/types";
import { getApiError } from "@/utils";
import { auth } from "@/auth";

export const deleteLinkAction = async (linkId: string) => {
  try {
    const session = await auth();
    const res = await remove<ApiResponse<void>>(`links/${linkId}`, session!);
    return res;
  } catch (error) {
    return getApiError(error);
  }
};
