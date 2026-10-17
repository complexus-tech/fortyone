import type {
  InfiniteData,
  QueryClient,
  QueryKey,
} from "@tanstack/react-query";
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

type MessagePages = InfiniteData<{ items: ChatMessage[] }>;

export const isChatMessageQuery = (key: QueryKey, workspaceSlug: string) =>
  key[0] === "chat" &&
  key[1] === workspaceSlug.toLowerCase() &&
  [
    "messages",
    "replies",
    "search",
    "latest",
    "message",
    "unread-threads",
  ].includes(String(key[2]));

export const getQueryMessages = (
  key: QueryKey,
  data: unknown,
): ChatMessage[] => {
  if (!data) return [];
  if (key[2] === "message") return [data as ChatMessage];
  if (key[2] === "latest") return (data as ChatMessagePage).items;
  return (data as MessagePages).pages.flatMap((page) => page.items);
};

export const getCachedMessage = (
  client: QueryClient,
  workspaceSlug: string,
  messageId: string,
) => {
  let latest: ChatMessage | undefined;
  for (const query of client.getQueryCache().findAll({
    predicate: ({ queryKey }: { queryKey: QueryKey }) =>
      isChatMessageQuery(queryKey, workspaceSlug),
  })) {
    for (const message of getQueryMessages(query.queryKey, query.state.data)) {
      if (
        message.id === messageId &&
        (!latest || message.version > latest.version)
      )
        latest = message;
    }
  }
  return latest;
};

/** Preserve page cursors and unread-thread metadata when updating a message. */
export const updateCachedMessage = (
  client: QueryClient,
  workspaceSlug: string,
  message: ChatMessage,
) => {
  client.setQueriesData<MessagePages>(
    {
      queryKey: chatKeys.all(workspaceSlug),
      predicate: (query) =>
        (["messages", "replies", "search"].includes(
          String(query.queryKey[2]),
        ) &&
          query.queryKey[3] === message.conversationId) ||
        query.queryKey[2] === "unread-threads",
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
          return { ...item, ...message };
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
  client.setQueryData<ChatMessage>(
    [...chatKeys.all(workspaceSlug), "message", message.id],
    (current) =>
      current && current.version > message.version
        ? current
        : { ...current, ...message },
  );
};
