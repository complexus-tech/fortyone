"use server";

import { auth } from "@/auth";
import { put } from "@/lib/http";
import type { ApiResponse } from "@/types";
import { getApiError } from "@/utils";
import type { UpdateAiChatPayload } from "../types";

export const updateAiChatAction = async (
  id: string,
  payload: UpdateAiChatPayload,
  workspaceSlug: string,
) => {
  try {
    const session = await auth();
    const ctx = { session: session!, workspaceSlug };
    const result = await put<UpdateAiChatPayload, ApiResponse<null>>(
      `chat-sessions/${id}`,
      payload,
      ctx,
    );
    return result;
  } catch (error) {
    return getApiError(error);
  }
};
