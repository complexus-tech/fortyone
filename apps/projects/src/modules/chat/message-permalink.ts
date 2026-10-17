import type { ChatMessage } from "./types";

/** Construct a shareable link without carrying compose, search, or theme state. */
export const getChatMessageLink = (
  message: Pick<ChatMessage, "id" | "conversationId" | "threadRootId">,
  chatPath: string,
  origin: string,
) => {
  const url = new URL(chatPath, origin);
  url.search = new URLSearchParams({
    conversation: message.conversationId,
    message: message.id,
  }).toString();
  url.hash = "";
  if (message.threadRootId) {
    url.searchParams.set("thread", message.threadRootId);
  }
  return url.toString();
};
