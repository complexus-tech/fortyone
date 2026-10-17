import type {
  InfiniteData,
  QueryClient,
  QueryKey,
} from "@tanstack/react-query";
import type {
  ChatConversation,
  ChatMarkUnreadResult,
  ChatMessage,
  ChatPage,
  ChatUnreadSummary,
  ChatUnreadThread,
} from "./types";
import { getQueryMessages, isChatMessageQuery } from "./cache";

type Scope = { workspaceSlug: string; accountId: string; message: ChatMessage };
type Pending = Scope & { abandoned: boolean };
type Snapshot = { key: QueryKey; data: unknown; projected: unknown };
type Pages<T> = InfiniteData<ChatPage<T>>;
type Journal = {
  operations: Pending[];
  snapshots: Map<string, Snapshot>;
  tails: Map<string, Promise<unknown>>;
  publishing: boolean;
  readRevision: number;
  unsubscribe: () => void;
};
const journals = new WeakMap<QueryClient, Journal>();
const scopeKey = ({ workspaceSlug, accountId, message }: Scope) =>
  `${workspaceSlug.toLowerCase()}:${accountId}:${message.conversationId}`;
const matchesScope = (key: QueryKey, scope: Scope) => {
  if (key[0] !== "chat" || key[1] !== scope.workspaceSlug.toLowerCase())
    return false;
  if (key[2] === "unread" || key[2] === "unread-threads")
    return key[3] === scope.accountId;
  return (
    key[2] === "conversations" ||
    (key[2] === "conversation" && key[3] === scope.message.conversationId) ||
    (isChatMessageQuery(key, scope.workspaceSlug) &&
      (key[2] === "message"
        ? key[3] === scope.message.id || key[3] === scope.message.threadRootId
        : key[3] === scope.message.conversationId))
  );
};
const mapPages = <T>(data: Pages<T>, update: (item: T) => T): Pages<T> => ({
  ...data,
  pages: data.pages.map((page) => ({ ...page, items: page.items.map(update) })),
});
const mapMessages = (
  key: QueryKey,
  data: unknown,
  update: (message: ChatMessage) => ChatMessage,
) => {
  if (key[2] === "message") return update(data as ChatMessage);
  if (key[2] === "latest") {
    const page = data as ChatPage<ChatMessage>;
    return { ...page, items: page.items.map(update) };
  }
  return mapPages(data as Pages<ChatMessage>, update);
};

/** Remove only our unchanged projection from another local cache edit. */
const unproject = (
  base: unknown,
  projected: unknown,
  incoming: unknown,
  field?: string,
): unknown => {
  if (base === projected || projected === undefined) return incoming;
  if (incoming === projected) return base;
  if (
    field &&
    [
      "unreadCount",
      "unreadRootCount",
      "unreadReplyCount",
      "totalUnread",
      "unreadThreads",
    ].includes(field) &&
    typeof base === "number" &&
    typeof projected === "number" &&
    typeof incoming === "number"
  )
    return Math.max(0, incoming - (projected - base));
  if (!base || !projected || !incoming || typeof incoming !== "object")
    return incoming;
  if (
    Array.isArray(base) &&
    Array.isArray(projected) &&
    Array.isArray(incoming)
  ) {
    const identity = (item: unknown) =>
      item && typeof item === "object"
        ? (item as { id?: string; conversationId?: string }).id ??
          (item as { conversationId?: string }).conversationId
        : undefined;
    return incoming.flatMap((item: unknown, index) => {
      const id = identity(item);
      const before = id
        ? base.find((entry: unknown) => identity(entry) === id)
        : base[index];
      const after = id
        ? projected.find((entry: unknown) => identity(entry) === id)
        : projected[index];
      // Inserted unread-thread rows are overlays even if a local reaction changed them.
      if (id && !before && after) return [];
      return [unproject(before, after, item)];
    });
  }
  if (
    Array.isArray(base) ||
    Array.isArray(projected) ||
    Array.isArray(incoming)
  )
    return incoming;
  const previous = base as Record<string, unknown>;
  const after = projected as Record<string, unknown>;
  return Object.fromEntries(
    Object.entries(incoming).map(([key, value]) => [
      key,
      unproject(previous[key], after[key], value, key),
    ]),
  );
};

const applyUnread = (
  snapshots: Map<string, Snapshot>,
  scope: Scope,
  response?: ChatMarkUnreadResult,
  preserveNewerCounts = false,
) => {
  const { message } = scope;
  const entries = Array.from(snapshots.values()).filter(
    ({ key, data }) => data !== undefined && matchesScope(key, scope),
  );
  const conversations: ChatConversation[] = [];
  const messages = new Map<string, ChatMessage>();
  let unreadThread: ChatUnreadThread | undefined;
  let summaryConversation:
    | ChatUnreadSummary["conversations"][number]
    | undefined;
  for (const { key, data } of entries) {
    if (key[2] === "conversation")
      conversations.unshift(data as ChatConversation);
    else if (key[2] === "conversations")
      for (const page of (data as Pages<ChatConversation>).pages)
        conversations.push(
          ...page.items.filter((item) => item.id === message.conversationId),
        );
    else if (key[2] === "unread")
      summaryConversation = (data as ChatUnreadSummary).conversations.find(
        (item) => item.conversationId === message.conversationId,
      );
    else if (isChatMessageQuery(key, scope.workspaceSlug)) {
      for (const item of getQueryMessages(key, data)) {
        if (item.conversationId !== message.conversationId) continue;
        const previous = messages.get(item.id);
        if (!previous || item.version >= previous.version)
          messages.set(item.id, item);
        if (key[2] === "unread-threads" && item.id === message.threadRootId)
          unreadThread = item as ChatUnreadThread;
      }
    }
  }
  messages.set(message.id, messages.get(message.id) ?? message);
  const conversation = conversations.at(0);
  const root = message.threadRootId
    ? messages.get(message.threadRootId)
    : undefined;
  const oldRootCount = Math.max(
    0,
    summaryConversation?.unreadRootCount ?? 0,
    ...conversations.map((item) => item.unreadRootCount),
  );
  const oldReplyCount = Math.max(
    0,
    summaryConversation?.unreadReplyCount ?? 0,
    ...conversations.map((item) => item.unreadReplyCount),
  );
  const oldThreadCount = Math.max(
    root?.unreadReplyCount ?? 0,
    unreadThread?.unreadReplyCount ?? 0,
  );
  const oldSequence = conversation?.lastReadSequence ?? 0;
  const eligible = Array.from(messages.values()).filter(
    (item) =>
      !item.deliveryStatus &&
      !item.deletedAt &&
      item.author.id !== scope.accountId &&
      item.threadRootId === message.threadRootId,
  );
  const knownUnread = eligible.filter(
    (item) => item.sequence >= message.sequence,
  ).length;
  const firstUnreadSequence = unreadThread
    ? messages.get(unreadThread.firstUnreadReplyId)?.sequence
    : undefined;
  const additional = eligible.filter(
    (item) =>
      item.sequence >= message.sequence &&
      item.sequence <
        (message.threadRootId
          ? firstUnreadSequence ?? message.sequence
          : oldSequence + 1),
  ).length;
  const serverCounts = response?.summary.conversations.find(
    (item) => item.conversationId === message.conversationId,
  );
  let responseCount = response?.unreadCount;
  if (response && preserveNewerCounts)
    responseCount = Math.max(
      response.unreadCount,
      message.threadRootId ? oldThreadCount : oldRootCount,
    );
  const threadCount = message.threadRootId
    ? responseCount ?? Math.max(knownUnread, oldThreadCount + additional)
    : oldThreadCount;
  const rootCount = message.threadRootId
    ? oldRootCount
    : responseCount ?? Math.max(oldRootCount + additional, knownUnread);
  let replyCount = oldReplyCount;
  if (message.threadRootId)
    replyCount =
      response && !preserveNewerCounts && serverCounts
        ? serverCounts.unreadReplyCount
        : Math.max(threadCount, oldReplyCount + threadCount - oldThreadCount);
  const unreadCount = rootCount + replyCount;
  const lastReadSequence = message.threadRootId
    ? oldSequence
    : Math.min(
        oldSequence,
        response?.lastReadSequence ?? Math.max(0, message.sequence - 1),
      );
  const firstUnreadReplyId =
    response?.firstUnreadReplyId ??
    (firstUnreadSequence !== undefined && firstUnreadSequence < message.sequence
      ? unreadThread?.firstUnreadReplyId
      : message.id);
  const patchConversation = (item: ChatConversation) =>
    item.id === message.conversationId
      ? {
          ...item,
          unreadCount,
          unreadRootCount: rootCount,
          unreadReplyCount: replyCount,
          ...(!message.threadRootId ? { lastReadSequence } : {}),
        }
      : item;
  const patchRoot = (item: ChatMessage) =>
    item.id === message.threadRootId
      ? { ...item, unreadReplyCount: threadCount }
      : item;
  for (const entry of entries) {
    const { key, data } = entry;
    if (key[2] === "conversation")
      entry.data = patchConversation(data as ChatConversation);
    else if (key[2] === "conversations")
      entry.data = mapPages(data as Pages<ChatConversation>, patchConversation);
    else if (key[2] === "unread") {
      const summary = data as ChatUnreadSummary;
      const before = summary.conversations.find(
        (item) => item.conversationId === message.conversationId,
      );
      const count = {
        conversationId: message.conversationId,
        unreadCount,
        unreadRootCount: rootCount,
        unreadReplyCount: replyCount,
      };
      entry.data = {
        ...summary,
        totalUnread: Math.max(
          0,
          summary.totalUnread + unreadCount - (before?.unreadCount ?? 0),
        ),
        unreadThreads:
          summary.unreadThreads +
          (message.threadRootId
            ? Number(threadCount > 0) - Number(oldThreadCount > 0)
            : 0),
        conversations: before
          ? summary.conversations.map((item) =>
              item.conversationId === message.conversationId ? count : item,
            )
          : [...summary.conversations, count],
      } satisfies ChatUnreadSummary;
    } else if (key[2] === "unread-threads" && message.threadRootId) {
      const pages = mapPages(data as Pages<ChatUnreadThread>, (item) =>
        item.id === message.threadRootId
          ? {
              ...item,
              unreadReplyCount: threadCount,
              firstUnreadReplyId: firstUnreadReplyId ?? item.firstUnreadReplyId,
            }
          : item,
      );
      if (
        !unreadThread &&
        root &&
        conversation &&
        pages.pages.length &&
        threadCount
      )
        pages.pages[0] = {
          ...pages.pages[0],
          items: [
            {
              ...root,
              unreadReplyCount: threadCount,
              firstUnreadReplyId: firstUnreadReplyId ?? message.id,
              conversationName: conversation.name,
              conversationKind: conversation.kind,
            },
            ...pages.pages[0].items,
          ],
        };
      if (!threadCount)
        pages.pages = pages.pages.map((page) => ({
          ...page,
          items: page.items.filter((item) => item.id !== message.threadRootId),
        }));
      entry.data = pages;
    } else if (
      message.threadRootId &&
      isChatMessageQuery(key, scope.workspaceSlug)
    )
      entry.data = mapMessages(key, data, patchRoot);
  }
};

const publish = (client: QueryClient, journal: Journal) => {
  for (const query of client.getQueryCache().findAll()) {
    if (
      query.state.data === undefined ||
      journal.snapshots.has(query.queryHash) ||
      !journal.operations.some(
        (operation) =>
          !operation.abandoned && matchesScope(query.queryKey, operation),
      )
    )
      continue;
    journal.snapshots.set(query.queryHash, {
      key: query.queryKey,
      data: query.state.data,
      projected: query.state.data,
    });
  }
  const projected = new Map(
    Array.from(journal.snapshots, ([hash, snapshot]) => [
      hash,
      { ...snapshot },
    ]),
  );
  for (const operation of journal.operations)
    if (!operation.abandoned) applyUnread(projected, operation);
  journal.publishing = true;
  try {
    for (const [hash, snapshot] of projected) {
      if (!client.getQueryCache().get(hash)) continue;
      client.setQueryData(snapshot.key, snapshot.data);
      journal.snapshots.get(hash)!.projected = client.getQueryData(
        snapshot.key,
      );
    }
  } finally {
    journal.publishing = false;
  }
  // Settled conversations no longer belong to this journal. Keeping their old
  // snapshots would overwrite a later cache update when another room settles.
  for (const [hash, snapshot] of journal.snapshots)
    if (
      !journal.operations.some(
        (operation) =>
          !operation.abandoned && matchesScope(snapshot.key, operation),
      )
    )
      journal.snapshots.delete(hash);
};
const getJournal = (client: QueryClient) => {
  const existing = journals.get(client);
  if (existing) return existing;
  const journal: Journal = {
    operations: [],
    snapshots: new Map(),
    tails: new Map(),
    publishing: false,
    readRevision: 0,
    unsubscribe: () => undefined,
  };
  journals.set(client, journal);
  journal.unsubscribe = client.getQueryCache().subscribe((event) => {
    if (journal.publishing) return;
    if (event.type === "removed") {
      journal.snapshots.delete(event.query.queryHash);
      for (const operation of journal.operations)
        if (
          !client
            .getQueryCache()
            .findAll()
            .some((query) => matchesScope(query.queryKey, operation))
        )
          operation.abandoned = true;
      return;
    }
    if (
      event.type !== "updated" ||
      event.action.type !== "success" ||
      !journal.operations.some(
        (operation) =>
          !operation.abandoned &&
          matchesScope(event.query.queryKey as QueryKey, operation),
      )
    )
      return;
    if (!event.action.manual) journal.readRevision += 1;
    const previous = journal.snapshots.get(event.query.queryHash);
    journal.snapshots.set(event.query.queryHash, {
      key: event.query.queryKey,
      data:
        previous && event.action.manual
          ? unproject(previous.data, previous.projected, event.query.state.data)
          : event.query.state.data,
      projected: event.query.state.data,
    });
    publish(client, journal);
  });
  return journal;
};

/** Rewind one stream immediately, without consuming unread messages in another. */
export const mutateUnreadOptimistically = (
  client: QueryClient,
  scope: Scope,
  execute: () => Promise<ChatMarkUnreadResult>,
): Promise<ChatMarkUnreadResult> => {
  const journal = getJournal(client);
  const operation: Pending = { ...scope, abandoned: false };
  const key = scopeKey(scope);
  const isAbandoned = () => operation.abandoned;
  const cancelled = client.cancelQueries(
    { predicate: (query) => matchesScope(query.queryKey, scope) },
    { revert: false },
  );
  journal.operations.push(operation);
  publish(client, journal);
  const delivery = (journal.tails.get(key) ?? Promise.resolve())
    .catch(() => undefined)
    .then(async () => {
      try {
        await cancelled;
        if (isAbandoned())
          throw new Error("This chat session is no longer active.");
        const revision = journal.readRevision;
        const response = await execute();
        if (!isAbandoned())
          applyUnread(
            journal.snapshots,
            scope,
            response,
            journal.readRevision !== revision,
          );
        return response;
      } finally {
        journal.operations = journal.operations.filter(
          (item) => item !== operation,
        );
        publish(client, journal);
        if (!journal.operations.length) {
          journal.unsubscribe();
          journals.delete(client);
        }
      }
    });
  journal.tails.set(key, delivery);
  return delivery;
};
