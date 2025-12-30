import type { Session } from "next-auth";
import { get } from "@/lib/http";
import type { ApiResponse } from "@/types";
import type { AiTotalChatMessages } from "../types";

export const getTotalMessagesForTheMonth = async (session: Session) => {
  const chats = await get<ApiResponse<AiTotalChatMessages>>(
    "chat-sessions/messages/count",
    session,
  );
  return chats?.data?.count || 0;
};
