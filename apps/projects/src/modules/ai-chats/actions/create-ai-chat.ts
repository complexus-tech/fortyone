"use server";

import { auth } from "@/auth";
import { post } from "@/lib/http";
import type { ApiResponse } from "@/types";
import { getApiError } from "@/utils";
import type { CreateAiChatPayload, AiChatSession } from "../types";

export const createAiChatAction = async (
  payload: CreateAiChatPayload,
  workspaceSlug: string,
) => {
  try {
    const session = await auth();
    const ctx = { session: session!, workspaceSlug };
    const chat = await post<CreateAiChatPayload, ApiResponse<AiChatSession>>(
      "chat-sessions",
      payload,
      ctx,
    );
    return chat;
  } catch (error) {
    return getApiError(error);
  }
};
