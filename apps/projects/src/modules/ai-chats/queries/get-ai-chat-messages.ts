import { get, type WorkspaceCtx } from "@/lib/http";
import type { ApiResponse } from "@/types";
import type { MayaUIMessage } from "@/lib/ai/tools/types";

export const getAiChatMessages = async (ctx: WorkspaceCtx, id: string) => {
  const messages = await get<ApiResponse<MayaUIMessage[]>>(
    `chat-sessions/${id}/messages`,
    ctx,
  );
  return messages.data!;
};
