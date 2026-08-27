import type { UIMessage } from "ai";
import type { RequestOptions } from "api-client";
import { auth } from "@/auth";
import { post } from "@/lib/http";
import type { ApiResponse } from "@/types";

export type MessageWriteOperation = "append" | "approval" | "regenerate";

export type MessageWriteReservation<
  UIMessageType extends UIMessage = UIMessage,
> = {
  generation: number;
  /**
   * Server-authoritative, request-safe history when durable approval receipts
   * repaired the submitted transcript. Historical file payloads remain
   * redacted, so this can safely replace the request copy for model/stream use.
   */
  messages?: UIMessageType[];
  token: string;
};

export type MessageWriteResult = {
  applied: boolean;
};

const MESSAGE_WRITE_REQUEST_TIMEOUT_MS = 10_000;

const getAuthenticatedWorkspaceContext = async (workspaceSlug: string) => {
  const session = await auth();
  if (!session?.user) {
    throw new Error("Chat persistence authentication is required.");
  }
  if (!workspaceSlug) {
    throw new Error("Chat persistence workspace is required.");
  }
  return { session, workspaceSlug };
};

const messageWriteOptions = {
  retry: 0,
  timeout: MESSAGE_WRITE_REQUEST_TIMEOUT_MS,
} satisfies RequestOptions;

const idempotentMessageWriteOptions = {
  retry: {
    limit: 1,
    methods: ["post"],
    retryOnTimeout: true,
    statusCodes: [408, 500, 502, 503, 504],
  },
  timeout: MESSAGE_WRITE_REQUEST_TIMEOUT_MS,
} satisfies RequestOptions;

export const beginAiChatMessageWrite = async <UIMessageType extends UIMessage>({
  chatId,
  messageId,
  messages,
  operation,
  title,
  workspaceSlug,
}: {
  chatId: string;
  messageId?: string;
  messages: UIMessageType[];
  operation: MessageWriteOperation;
  title: string;
  workspaceSlug: string;
}): Promise<MessageWriteReservation<UIMessageType>> => {
  const ctx = await getAuthenticatedWorkspaceContext(workspaceSlug);
  const response = await post<
    {
      messages: UIMessage[];
      messageId?: string;
      operation: MessageWriteOperation;
      title: string;
    },
    ApiResponse<MessageWriteReservation<UIMessageType>>
  >(
    `chat-sessions/${encodeURIComponent(chatId)}/message-writes/begin`,
    { messageId, messages, operation, title },
    ctx,
    messageWriteOptions,
  );
  if (!response.data?.token || response.data.generation <= 0) {
    throw new Error("Chat write reservation returned an invalid token.");
  }
  return response.data;
};

export const finalizeAiChatMessageWrite = async <
  UIMessageType extends UIMessage,
>({
  chatId,
  messages,
  reservation,
  workspaceSlug,
}: {
  chatId: string;
  messages: UIMessageType[];
  reservation: MessageWriteReservation<UIMessageType>;
  workspaceSlug: string;
}): Promise<MessageWriteResult> => {
  const ctx = await getAuthenticatedWorkspaceContext(workspaceSlug);
  const response = await post<
    {
      generation: number;
      messages: UIMessage[];
      token: string;
    },
    ApiResponse<MessageWriteResult>
  >(
    `chat-sessions/${encodeURIComponent(chatId)}/message-writes/finalize`,
    {
      generation: reservation.generation,
      messages,
      token: reservation.token,
    },
    ctx,
    idempotentMessageWriteOptions,
  );
  if (!response.data || typeof response.data.applied !== "boolean") {
    throw new Error("Chat write finalization returned no result.");
  }
  return response.data;
};

export const recoverMutationApprovalOutput = async ({
  chatId,
  fingerprint,
  toolCallId,
  workspaceSlug,
}: {
  chatId: string;
  fingerprint: string;
  toolCallId: string;
  workspaceSlug: string;
}): Promise<MessageWriteResult> => {
  const ctx = await getAuthenticatedWorkspaceContext(workspaceSlug);
  const response = await post<
    { fingerprint: string },
    ApiResponse<MessageWriteResult>
  >(
    `chat-sessions/${encodeURIComponent(chatId)}/mutation-approvals/${encodeURIComponent(toolCallId)}/recover-output`,
    { fingerprint },
    ctx,
    idempotentMessageWriteOptions,
  );
  if (!response.data || typeof response.data.applied !== "boolean") {
    throw new Error("Mutation approval recovery returned no result.");
  }
  return response.data;
};
