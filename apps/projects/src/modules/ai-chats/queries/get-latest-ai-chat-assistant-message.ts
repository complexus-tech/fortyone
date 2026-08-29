import { get, type WorkspaceCtx } from "@/lib/http";
import type { MayaUIMessage } from "@/lib/ai/tools/types";
import type { ApiResponse } from "@/types";

export const getLatestAiChatAssistantMessage = async (
  ctx: WorkspaceCtx,
  id: string,
) => {
  const response = await get<ApiResponse<MayaUIMessage | null>>(
    `chat-sessions/${id}/messages/latest-assistant`,
    ctx,
  );
  return response.data ?? null;
};
