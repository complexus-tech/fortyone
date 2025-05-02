"use server";

import { auth } from "@/auth";
import { remove } from "@/lib/http";
import type { ApiResponse } from "@/types";
import { getApiError } from "@/utils";

export const deleteStateAction = async (stateId: string) => {
  try {
    const session = await auth();
    const res = await remove<ApiResponse<void>>(`states/${stateId}`, session!);
    return res;
  } catch (error) {
    return getApiError(error);
  }
};
