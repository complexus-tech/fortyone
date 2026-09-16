import { safeValidateUIMessages } from "ai";
import { get, remove } from "@/lib/http";
import type { ApiResponse } from "@/types";
import type { MayaMessage } from "../types";
import { assertChatId, parseMayaSessions } from "./chat-protocol";

export const getMayaSessions = async (userId: string, signal: AbortSignal) => {
  const response = await get<ApiResponse<unknown>>("chat-sessions", { signal });
  return parseMayaSessions(response.data, userId);
};

export const validateMayaMessages = async (
  messages: unknown,
): Promise<MayaMessage[]> => {
  const result = await safeValidateUIMessages({ messages });
  if (!result.success)
    throw new Error("Maya returned invalid conversation messages.");
  return result.data;
};

export const getMayaMessages = async (id: string, signal: AbortSignal) => {
  const response = await get<ApiResponse<unknown>>(
    `chat-sessions/${assertChatId(id)}/messages`,
    { signal },
  );
  return validateMayaMessages(response.data);
};

export const deleteMayaSession = (id: string, signal: AbortSignal) =>
  remove(`chat-sessions/${assertChatId(id)}`, { signal });
