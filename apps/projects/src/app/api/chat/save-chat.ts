import type { UIMessage } from "ai";
import {
  beginAiChatMessageWrite,
  finalizeAiChatMessageWrite,
} from "@/modules/ai-chats/actions/message-write";
import type {
  MessageWriteOperation,
  MessageWriteReservation,
  MessageWriteResult,
} from "@/modules/ai-chats/actions/message-write";
import { getChatTitle } from "./chat-title";

const SLOW_MESSAGE_WRITE_MS = 2_000;

const getPersistenceDiagnostic = (error: unknown) => {
  const errorCode =
    error &&
    typeof error === "object" &&
    "code" in error &&
    typeof error.code === "string"
      ? error.code
      : "unknown";
  return {
    errorCode,
    errorType: error instanceof Error ? error.name : typeof error,
  };
};

export const beginChatWrite = async <UIMessageType extends UIMessage>({
  id,
  messageId,
  messages,
  operation,
  workspaceSlug,
}: {
  id: string;
  messageId?: string;
  messages: UIMessageType[];
  operation: MessageWriteOperation;
  workspaceSlug: string;
}): Promise<MessageWriteReservation<UIMessageType>> => {
  const startedAt = Date.now();
  try {
    const reservation = await beginAiChatMessageWrite({
      chatId: id,
      messageId,
      messages,
      operation,
      title: getChatTitle(messages),
      workspaceSlug,
    });
    const durationMs = Date.now() - startedAt;
    if (durationMs >= SLOW_MESSAGE_WRITE_MS) {
      // eslint-disable-next-line no-console -- Payload-free latency diagnostics keep the extra persistence call observable.
      console.warn("[chat/save] Slow Maya write reservation", {
        chatId: id,
        durationMs,
        operation,
        workspaceSlug,
      });
    }
    return reservation;
  } catch (error) {
    // eslint-disable-next-line no-console -- Never log transcript content or approval inputs.
    console.error("[chat/save] Failed to reserve Maya conversation write", {
      chatId: id,
      ...getPersistenceDiagnostic(error),
      operation,
      workspaceSlug,
    });
    throw error;
  }
};

export const saveChat = async <UIMessageType extends UIMessage>({
  id,
  messages,
  reservation,
  workspaceSlug,
}: {
  id: string;
  messages: UIMessageType[];
  reservation: MessageWriteReservation<UIMessageType>;
  workspaceSlug: string;
}): Promise<MessageWriteResult> => {
  try {
    return await finalizeAiChatMessageWrite({
      chatId: id,
      messages,
      reservation,
      workspaceSlug,
    });
  } catch (error) {
    // eslint-disable-next-line no-console -- Persisting the conversation must remain observable without exposing message content.
    console.error("[chat/save] Failed to finalize Maya conversation write", {
      chatId: id,
      ...getPersistenceDiagnostic(error),
      generation: reservation.generation,
      workspaceSlug,
    });
    throw error;
  }
};
