import type { Session } from "next-auth";
import { get } from "@/lib/http";
import type { ApiResponse } from "@/types";
import type { AiChatSession } from "../types";

export const getAiChats = async (session: Session) => {
  const chats = await get<ApiResponse<AiChatSession[]>>(
    "chat-sessions",
    session,
  );
  return chats.data!;
};
