import type { Message } from "ai";
import type { Session } from "next-auth";
import { get } from "@/lib/http";
import type { ApiResponse } from "@/types";

export const getAiChatMessages = async (session: Session, id: string) => {
  const messages = await get<ApiResponse<Message[]>>(
    `chat-sessions/${id}/messages`,
    session,
  );
  return messages.data!;
};
