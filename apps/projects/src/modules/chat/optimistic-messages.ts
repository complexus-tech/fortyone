import type { QueryClient, QueryKey } from "@tanstack/react-query";
import type { ChatMessage, ChatPerson } from "./types";
import {
  getCachedMessage,
  getQueryMessages,
  isChatMessageQuery,
  updateCachedMessage,
} from "./cache";

type Operation = {
  apply: (message: ChatMessage) => ChatMessage;
  commit?: (
    current: ChatMessage,
    response: ChatMessage,
    refreshed: boolean,
  ) => ChatMessage;
  execute: (current: ChatMessage) => Promise<ChatMessage>;
};
type PendingMessage = {
  workspace: string;
  confirmed: ChatMessage;
  operations: Operation[];
  published: WeakSet<ChatMessage>;
  tail: Promise<unknown>;
  abandoned: boolean;
  revision: number;
};
type Journal = {
  messages: Map<string, PendingMessage>;
  publishing: boolean;
  unsubscribe: () => void;
};
const journals = new WeakMap<QueryClient, Journal>();
const messageKey = (workspace: string, id: string) =>
  `${workspace.toLowerCase()}:${id}`;

const publish = (
  client: QueryClient,
  journal: Journal,
  entry: PendingMessage,
) => {
  if (entry.abandoned) return;
  const projected = entry.operations.reduce(
    (message, operation) => operation.apply(message),
    entry.confirmed,
  );
  journal.publishing = true;
  try {
    updateCachedMessage(client, entry.workspace, projected);
    // React Query structurally shares objects, so remember the actual cached
    // instances, not just the object passed to setQueryData.
    for (const query of client.getQueryCache().findAll({
      predicate: ({ queryKey }: { queryKey: QueryKey }) =>
        isChatMessageQuery(queryKey, entry.workspace),
    })) {
      for (const message of getQueryMessages(
        query.queryKey,
        query.state.data,
      )) {
        if (message.id === projected.id) entry.published.add(message);
      }
    }
  } finally {
    journal.publishing = false;
  }
};

const getJournal = (client: QueryClient) => {
  const current = journals.get(client);
  if (current) return current;
  const journal: Journal = {
    messages: new Map(),
    publishing: false,
    unsubscribe: () => undefined,
  };
  journals.set(client, journal);
  journal.unsubscribe = client.getQueryCache().subscribe((event) => {
    if (event.type === "removed") {
      // Clearing account caches must not be undone by a late mutation response.
      // Ordinary garbage collection of one duplicate cache is harmless.
      for (const entry of journal.messages.values()) {
        if (!getCachedMessage(client, entry.workspace, entry.confirmed.id))
          entry.abandoned = true;
      }
      return;
    }
    if (
      journal.publishing ||
      event.type !== "updated" ||
      event.action.type !== "success"
    )
      return;
    const { state } = event.query;
    const queryKey = event.query.queryKey as QueryKey;
    const workspace = queryKey[1];
    if (
      typeof workspace !== "string" ||
      !isChatMessageQuery(queryKey, workspace)
    )
      return;
    for (const message of getQueryMessages(queryKey, state.data)) {
      const entry = journal.messages.get(messageKey(workspace, message.id));
      if (!entry || entry.published.has(message)) continue;
      if (message.version >= entry.confirmed.version) {
        entry.confirmed = message;
        entry.revision += 1;
      }
      // SSE refreshes and pagination may finish while a write is in flight.
      // Rebase local intentions on that response instead of letting it erase them.
      publish(client, journal, entry);
    }
  });
  return journal;
};

/**
 * Keep writes ordered per message, but apply every intention immediately.
 * Failed operations are removed individually; no snapshot rollback can erase
 * a different reaction, a new reply, or an authorized realtime refresh.
 */
export const mutateMessageOptimistically = (
  client: QueryClient,
  workspaceSlug: string,
  original: ChatMessage,
  operation: Operation,
): Promise<ChatMessage> => {
  const journal = getJournal(client);
  const key = messageKey(workspaceSlug, original.id);
  let entry = journal.messages.get(key);
  if (!entry) {
    entry = {
      workspace: workspaceSlug,
      confirmed:
        getCachedMessage(client, workspaceSlug, original.id) ?? original,
      operations: [],
      published: new WeakSet(),
      tail: Promise.resolve(),
      abandoned: false,
      revision: 0,
    };
    journal.messages.set(key, entry);
  }
  const pending = entry;
  const cancelReads = () =>
    client.cancelQueries(
      {
        predicate: ({ queryKey }) =>
          isChatMessageQuery(queryKey, workspaceSlug) &&
          (queryKey[2] === "unread-threads" ||
            queryKey[3] === original.conversationId ||
            (queryKey[2] === "message" && queryKey[3] === original.id)),
      },
      { revert: false },
    );
  const canceled = cancelReads();
  pending.operations.push(operation);
  publish(client, journal, pending);

  const delivery = pending.tail
    .catch(() => undefined)
    .then(async () => {
      await canceled;
      try {
        if (pending.abandoned)
          throw new Error("This chat session is no longer active.");
        const revision = pending.revision;
        const response = await operation.execute(pending.confirmed);
        await cancelReads();
        if (response.version >= pending.confirmed.version) {
          pending.confirmed = operation.commit
            ? operation.commit(
                pending.confirmed,
                response,
                pending.revision !== revision,
              )
            : response;
        }
        return response;
      } finally {
        pending.operations = pending.operations.filter(
          (item) => item !== operation,
        );
        publish(client, journal, pending);
        if (!pending.operations.length) {
          journal.messages.delete(key);
          if (!journal.messages.size) {
            journal.unsubscribe();
            journals.delete(client);
          }
        }
      }
    });
  pending.tail = delivery;
  return delivery;
};

export const setMessageReaction = (
  message: ChatMessage,
  emoji: string,
  selected: boolean,
  person: ChatPerson,
): ChatMessage => {
  if (message.deletedAt) return message;
  const existing = message.reactions.find(
    (reaction) => reaction.emoji === emoji,
  );
  if (Boolean(existing?.reacted) === selected) return message;
  const count = Math.max(0, (existing?.count ?? 0) + (selected ? 1 : -1));
  const users = (existing?.users ?? []).filter((user) => user.id !== person.id);
  if (selected) users.push(person);
  const reactions = message.reactions.filter(
    (reaction) => reaction.emoji !== emoji,
  );
  if (count) {
    const reaction = { emoji, count, reacted: selected, users };
    const index = message.reactions.findIndex((item) => item.emoji === emoji);
    reactions.splice(index < 0 ? reactions.length : index, 0, reaction);
  }
  return { ...message, reactions };
};

export const commitMessageReaction = (
  current: ChatMessage,
  response: ChatMessage,
  emoji: string,
  selected: boolean,
  person: ChatPerson,
  refreshed: boolean,
): ChatMessage => {
  // Reactions have no server revision of their own. A response captured before
  // an intervening SSE read must not discard peer reactions from that read.
  if (refreshed) return setMessageReaction(current, emoji, selected, person);
  const reaction = response.reactions.find((item) => item.emoji === emoji);
  const reactions = current.reactions.filter((item) => item.emoji !== emoji);
  if (reaction) {
    const index = current.reactions.findIndex((item) => item.emoji === emoji);
    reactions.splice(index < 0 ? reactions.length : index, 0, reaction);
  }
  return { ...current, reactions };
};

export const deletedMessage = (
  message: ChatMessage,
  deletedAt: string,
): ChatMessage => ({
  ...message,
  deletedAt,
  text: "",
  richContent: null,
  reactions: [],
  attachments: [],
  attachmentIds: [],
  links: [],
  taskRefs: [],
  mentionedUserIds: [],
  agentMentions: [],
});
