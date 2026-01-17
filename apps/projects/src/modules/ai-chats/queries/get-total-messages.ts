import { get, type WorkspaceCtx } from "@/lib/http";
import type { ApiResponse } from "@/types";
import type { AiTotalChatMessages } from "../types";

export const getTotalMessagesForTheMonth = async (ctx: WorkspaceCtx) => {
  const chats = await get<ApiResponse<AiTotalChatMessages>>(
    "chat-sessions/messages/count",
    ctx,
  );
  return chats?.data?.count || 0;
};
