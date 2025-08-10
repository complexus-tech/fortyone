import type { Session } from "next-auth";
import { get } from "@/lib/http";
import type { ApiResponse } from "@/types";
import type { MayaUIMessage } from "@/lib/ai/tools/types";

export const getAiChatMessages = async (session: Session, id: string) => {
  const messages = await get<ApiResponse<MayaUIMessage[]>>(
    `chat-sessions/${id}/messages`,
    session,
  );
  return messages.data!;
};
