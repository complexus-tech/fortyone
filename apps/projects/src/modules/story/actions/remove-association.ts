"use server";

import { remove } from "@/lib/http/fetch";
import type { ApiResponse } from "@/types";
import { getApiError } from "@/utils";
import { auth } from "@/auth";

export const removeAssociationAction = async (
  associationId: string,
) => {
  try {
    const session = await auth();
    const res = await remove<ApiResponse<null>>(
      `stories/associations/${associationId}`,
      session!,
    );
    return res;
  } catch (error) {
    return getApiError(error);
  }
};
