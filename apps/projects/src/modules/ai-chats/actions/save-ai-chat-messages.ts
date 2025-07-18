"use server";

import { auth } from "@/auth";
import { post } from "@/lib/http";
import type { ApiResponse } from "@/types";
import { getApiError } from "@/utils";
import type { SaveMessagesPayload } from "../types";

export const saveAiChatMessagesAction = async (
  payload: SaveMessagesPayload,
) => {
  try {
    const session = await auth();
    const result = await post<SaveMessagesPayload, ApiResponse<null>>(
      `chat-sessions/${payload.id}/messages`,
      payload,
      session!,
    );
    return result;
  } catch (error) {
    return getApiError(error);
  }
};
