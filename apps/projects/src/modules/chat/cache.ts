import type { InfiniteData, QueryClient } from "@tanstack/react-query";
import type { ChatMessage, ChatMessagePage } from "./types";
import { chatKeys } from "./constants";

export const flattenMessages = (
  pages: readonly ChatMessagePage[] | undefined,
): ChatMessage[] => {
  const messages = new Map<string, ChatMessage>();
  for (const page of pages ?? []) {
    for (const message of page.items) {
      const previous = messages.get(message.id);
      if (!previous || message.version >= previous.version)
        messages.set(message.id, message);
    }
  }
  return Array.from(messages.values()).sort((a, b) => a.sequence - b.sequence);
};

export const updateCachedMessage = (
  client: QueryClient,
  workspaceSlug: string,
  message: ChatMessage,
) => {
  client.setQueriesData<InfiniteData<ChatMessagePage>>(
    {
      queryKey: chatKeys.all(workspaceSlug),
      predicate: (query) =>
        ["messages", "replies", "search"].includes(String(query.queryKey[2])) &&
        query.queryKey[3] === message.conversationId,
    },
    (current) => {
      if (!current) return current;
      const containsMessage = current.pages.some((page) =>
        page.items.some(
          (item) => item.id === message.id && item.version <= message.version,
        ),
      );
      if (!containsMessage) return current;
      const pages = current.pages.map((page) => ({
        ...page,
        items: page.items.map((item) => {
          if (item.id !== message.id || item.version > message.version)
            return item;
          return message;
        }),
      }));
      return { ...current, pages };
    },
  );
  client.setQueryData<ChatMessagePage>(
    [...chatKeys.all(workspaceSlug), "latest", message.conversationId],
    (current) =>
      current
        ? {
            ...current,
            items: current.items.map((item) =>
              item.id === message.id && item.version <= message.version
                ? message
                : item,
            ),
          }
        : current,
  );
  client.setQueryData(
    [...chatKeys.all(workspaceSlug), "message", message.id],
    message,
  );
};
