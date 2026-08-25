import type { ModelMessage, UIMessage } from "ai";
import { pruneMessages } from "ai";
import { compactToolOutput } from "@/lib/ai/compact-tool-output";

export const MAX_CHAT_CONTEXT_MESSAGES = 18;

export const selectRecentChatMessages = (messages: UIMessage[]) => {
  const recentMessages = messages.slice(-MAX_CHAT_CONTEXT_MESSAGES);
  const firstUserMessageIndex = recentMessages.findIndex(
    (message) => message.role === "user",
  );

  return firstUserMessageIndex > 0
    ? recentMessages.slice(firstUserMessageIndex)
    : recentMessages;
};

export const compactChatToolOutputs = (messages: UIMessage[]) =>
  messages.map(
    (message): UIMessage => ({
      ...message,
      parts: message.parts.map((part) => {
        if (!part.type.startsWith("tool-") || !("output" in part)) return part;

        return {
          ...part,
          output: compactToolOutput(part.output),
        } as typeof part;
      }),
    }),
  );

export const pruneChatModelMessages = (messages: ModelMessage[]) =>
  pruneMessages({
    messages,
    reasoning: "all",
    toolCalls: "before-last-6-messages",
    emptyMessages: "remove",
  });
