"use server";

import { auth } from "@/auth";
import { post } from "@/lib/http";
import type { ApiResponse } from "@/types";
import { getApiError } from "@/utils";
import type { SaveMessagesPayload } from "../types";

export const saveAiChatMessagesAction = async (
  payload: SaveMessagesPayload,
  workspaceSlug: string,
) => {
  try {
    const session = await auth();
    const ctx = { session: session!, workspaceSlug };
    const result = await post<SaveMessagesPayload, ApiResponse<null>>(
      `chat-sessions/${payload.id}/messages`,
      payload,
      ctx,
    );
    return result;
  } catch (error) {
    return getApiError(error);
  }
};
