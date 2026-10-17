import type { ChatMessage } from "./types";

/** Client-only metadata. It is never sent to or persisted by the API. */
export type OutgoingChatMessage = ChatMessage & {
  optimisticReplySummary?: { baseReplyCount: number; settled?: boolean };
  acknowledged?: boolean;
};

/**
 * Project pending replies without changing the authoritative parent cache.
 * The count is a floor: an SSE refresh may already include a pending send, so
 * adding one to every new server count would count that same reply twice.
 */
export const withOptimisticReplySummary = (
  root: ChatMessage,
  outgoing: readonly OutgoingChatMessage[],
): ChatMessage => {
  const replies = outgoing.filter(
    (message) =>
      message.conversationId === root.conversationId &&
      message.threadRootId === root.id &&
      message.optimisticReplySummary &&
      message.deliveryStatus !== "failed",
  );
  if (!replies.length) return root;

  const expectedReplyCount = Math.max(
    ...replies.map(
      (reply, index) =>
        reply.optimisticReplySummary!.baseReplyCount + replies.length - index,
    ),
  );
  const newest = [...replies].sort((a, b) =>
    b.createdAt.localeCompare(a.createdAt),
  );
  const authors = [
    ...new Map(
      [...newest.map((reply) => reply.author), ...root.replyAuthors].map(
        (author) => [author.id, author],
      ),
    ).values(),
  ].slice(0, 3);
  const lastReplyAt = [
    root.lastReplyAt,
    ...replies.map((reply) => reply.createdAt),
  ]
    .filter((time): time is string => Boolean(time))
    .sort()
    .at(-1)!;

  return {
    ...root,
    replyCount: Math.max(root.replyCount, expectedReplyCount),
    replyAuthors: authors,
    lastReplyAt,
  };
};
