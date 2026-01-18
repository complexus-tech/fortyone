import { get, type WorkspaceCtx } from "@/lib/http";
import type { ApiResponse } from "@/types";
import type { AiChatSession } from "../types";

export const getAiChats = async (ctx: WorkspaceCtx) => {
  const chats = await get<ApiResponse<AiChatSession[]>>("chat-sessions", ctx);
  return chats.data!;
};
