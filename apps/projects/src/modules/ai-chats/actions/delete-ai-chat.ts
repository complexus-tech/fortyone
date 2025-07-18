"use server";

import { auth } from "@/auth";
import { remove } from "@/lib/http";
import type { ApiResponse } from "@/types";
import { getApiError } from "@/utils";

export const deleteAiChatAction = async (id: string) => {
  try {
    const session = await auth();
    const result = await remove<ApiResponse<null>>(
      `chat-sessions/${id}`,
      session!,
    );
    return result;
  } catch (error) {
    return getApiError(error);
  }
};
