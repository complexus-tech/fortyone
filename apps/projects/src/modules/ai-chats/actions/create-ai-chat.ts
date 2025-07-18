"use server";

import { auth } from "@/auth";
import { post } from "@/lib/http";
import type { ApiResponse } from "@/types";
import { getApiError } from "@/utils";
import type { CreateAiChatPayload, AiChatSession } from "../types";

export const createAiChatAction = async (payload: CreateAiChatPayload) => {
  try {
    const session = await auth();
    const chat = await post<CreateAiChatPayload, ApiResponse<AiChatSession>>(
      "chat-sessions",
      payload,
      session!,
    );
    return chat;
  } catch (error) {
    return getApiError(error);
  }
};
