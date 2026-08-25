import type { ModelMessage, UIMessage } from "ai";
import { pruneMessages } from "ai";

export const MAX_CHAT_CONTEXT_MESSAGES = 24;

export const selectRecentChatMessages = (messages: UIMessage[]) => {
  const recentMessages = messages.slice(-MAX_CHAT_CONTEXT_MESSAGES);
  const firstUserMessageIndex = recentMessages.findIndex(
    (message) => message.role === "user",
  );

  return firstUserMessageIndex > 0
    ? recentMessages.slice(firstUserMessageIndex)
    : recentMessages;
};

export const pruneChatModelMessages = (messages: ModelMessage[]) =>
  pruneMessages({
    messages,
    reasoning: "all",
    toolCalls: "before-last-8-messages",
    emptyMessages: "remove",
  });
