import "server-only";

import { safeValidateUIMessages } from "ai";
import type { NextRequest } from "next/server";
import type { MayaUIMessage } from "@/lib/ai/tools/types";
import type { Memory } from "@/modules/ai-chats/types";
import type { getUserContext } from "./user-context";

const MAX_CHAT_REQUEST_BYTES = 16 * 1024 * 1024;
const INVALID_CHAT_REQUEST_CODE = "invalid_chat_request";

export type ChatRequestBody = {
  currentPath: string;
  currentTheme: string;
  id: string;
  memories: Memory[];
  messageId?: string;
  messages: MayaUIMessage[];
  provider?: "google" | "openai";
  resolvedTheme: string;
  subscription?: {
    billingEndsAt: string;
    billingInterval: string;
    status: string;
    tier: string;
    username?: string;
  };
  terminology: {
    keyResults: string;
    objectives: string;
    sprints: string;
    stories: string;
  };
  totalMessages: { current: number; limit: number };
  trigger?: "regenerate-message" | "submit-message";
  username?: string;
  workspace: Parameters<typeof getUserContext>[0]["workspace"];
};

const createInvalidChatRequestError = (cause?: unknown) =>
  Object.assign(new Error("Invalid Maya chat request."), {
    cause,
    code: INVALID_CHAT_REQUEST_CODE,
  });

export const isInvalidChatRequestError = (error: unknown) =>
  Boolean(
    error &&
      typeof error === "object" &&
      "code" in error &&
      error.code === INVALID_CHAT_REQUEST_CODE,
  );

const decodeChatRequestBody = async (body: unknown) => {
  if (!body || typeof body !== "object" || Array.isArray(body)) {
    throw createInvalidChatRequestError();
  }

  const messages = await safeValidateUIMessages<MayaUIMessage>({
    messages: Reflect.get(body, "messages"),
  });
  if (!messages.success) {
    throw createInvalidChatRequestError(messages.error);
  }

  return {
    ...(body as Omit<ChatRequestBody, "messages">),
    messages: messages.data,
  };
};

export const parseChatRequestBody = async (
  request: Pick<NextRequest, "headers" | "text">,
): Promise<ChatRequestBody> => {
  const declaredLength = Number(request.headers.get("content-length"));
  if (
    Number.isFinite(declaredLength) &&
    declaredLength > MAX_CHAT_REQUEST_BYTES
  ) {
    throw Object.assign(new Error("Maya chat request is too large."), {
      code: "request_too_large",
    });
  }

  const requestText = await request.text();
  if (
    new TextEncoder().encode(requestText).byteLength > MAX_CHAT_REQUEST_BYTES
  ) {
    throw Object.assign(new Error("Maya chat request is too large."), {
      code: "request_too_large",
    });
  }

  let body: unknown;
  try {
    body = JSON.parse(requestText) as unknown;
  } catch (error) {
    throw createInvalidChatRequestError(error);
  }

  return decodeChatRequestBody(body);
};

export const dispatchValidatedChatRequest = async ({
  handle,
  request,
}: {
  handle: (requestBody: ChatRequestBody) => Promise<Response> | Response;
  request: Pick<NextRequest, "headers" | "text">;
}) => {
  try {
    return await handle(await parseChatRequestBody(request));
  } catch (error) {
    if (isInvalidChatRequestError(error)) {
      return new Response("Invalid Maya chat request.", { status: 400 });
    }
    throw error;
  }
};
