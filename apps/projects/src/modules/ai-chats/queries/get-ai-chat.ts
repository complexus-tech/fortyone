import { get, type WorkspaceCtx } from "@/lib/http";
import type { ApiResponse } from "@/types";
import type { AiChatSession } from "../types";

export const getAiChat = async (ctx: WorkspaceCtx, id: string) => {
  const chat = await get<ApiResponse<AiChatSession>>(`chat-sessions/${id}`, ctx);
  return chat.data!;
};
