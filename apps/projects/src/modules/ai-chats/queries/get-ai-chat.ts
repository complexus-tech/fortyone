import type { Session } from "next-auth";
import { get } from "@/lib/http";
import type { ApiResponse } from "@/types";
import type { AiChatSession } from "../types";

export const getAiChat = async (session: Session, id: string) => {
  const chat = await get<ApiResponse<AiChatSession>>(
    `chat-sessions/${id}`,
    session,
  );
  return chat.data!;
};
