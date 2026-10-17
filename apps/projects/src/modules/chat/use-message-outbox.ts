"use client";

import { useCallback, useEffect, useRef } from "react";
import { skipToken, useQuery, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { useWorkspacePath } from "@/hooks/use-workspace-path";
import { useSession } from "@/lib/auth/client";
import type { ChatDraft } from "./chat-rich-text";
import type { ChatMessage } from "./types";
import * as api from "./api";
import { getMessageInput } from "./message-input";
import { extractMessageLinks } from "./chat-links";
import { updateCachedMessage } from "./cache";
import { reconcileChatQueries } from "./realtime";

const EMPTY_OUTBOX: ChatMessage[] = [];
const outboxKey = (workspaceSlug: string, accountId?: string) =>
  ["chat-outbox", workspaceSlug.toLowerCase(), accountId] as const;

/** Keep pending sends separate from HTTP pages so refetches cannot erase them. */
export const mergeOutgoingMessages = (
  messages: ChatMessage[],
  outgoing: ChatMessage[],
) => {
  const confirmed = new Set(
    messages.map((message) => `${message.author.id}:${message.creationId}`),
  );
  const messageIds = new Set(messages.map((message) => message.id));
  const remaining = outgoing.filter(
    (message) =>
      !confirmed.has(`${message.author.id}:${message.creationId}`) &&
      !messageIds.has(message.id),
  );
  const delivered = remaining.filter((message) => !message.deliveryStatus);
  const pending = remaining.filter((message) => message.deliveryStatus);
  return [...messages, ...delivered]
    .sort((a, b) => a.sequence - b.sequence)
    .concat(pending);
};

export const useMessageOutbox = () => {
  const client = useQueryClient();
  const { workspaceSlug } = useWorkspacePath();
  const { data: session } = useSession();
  const accountId = session?.user.id;
  const activeScope = useRef<object | null>(null);
  useEffect(() => {
    activeScope.current = {};
    return () => {
      activeScope.current = null;
    };
  }, [accountId, workspaceSlug]);
  const { data: outgoing = EMPTY_OUTBOX } = useQuery<ChatMessage[]>({
    queryKey: outboxKey(workspaceSlug, accountId),
    queryFn: skipToken,
    initialData: EMPTY_OUTBOX,
    enabled: false,
    gcTime: Infinity,
  });
  const update = useCallback(
    (transform: (messages: ChatMessage[]) => ChatMessage[]) => {
      client.setQueryData<ChatMessage[]>(
        outboxKey(workspaceSlug, accountId),
        (current = EMPTY_OUTBOX) => transform(current),
      );
    },
    [accountId, client, workspaceSlug],
  );
  const deliver = useCallback(
    async (pending: ChatMessage) => {
      const scope = activeScope.current;
      const isCurrent = () => scope !== null && activeScope.current === scope;
      try {
        const message = await api.createMessage(
          { workspaceSlug },
          pending.conversationId,
          {
            ...getMessageInput({
              text: pending.text,
              content: pending.richContent,
              attachments: pending.attachments,
            }),
            creationId: pending.creationId,
            threadRootId: pending.threadRootId ?? undefined,
          },
        );
        update((current) =>
          current.map((item) =>
            item.creationId === pending.creationId ? message : item,
          ),
        );
        // The account-scoped outbox survives navigation, but a late response
        // must not populate shared caches or interrupt a different session.
        if (!isCurrent()) return undefined;
        updateCachedMessage(client, workspaceSlug, message);
        void reconcileChatQueries(
          client,
          workspaceSlug,
          message.conversationId,
        ).catch((error: unknown) => {
          if (!isCurrent()) return;
          toast.error("Message sent. Could not refresh the conversation.", {
            description: error instanceof Error ? error.message : undefined,
          });
        });
        return message;
      } catch (error) {
        // A realtime read may already have confirmed this send before HTTP failed.
        const stillPending = client
          .getQueryData<ChatMessage[]>(outboxKey(workspaceSlug, accountId))
          ?.some((message) => message.creationId === pending.creationId);
        if (!stillPending) return undefined;
        update((current) =>
          current.map((item) =>
            item.creationId === pending.creationId
              ? { ...item, deliveryStatus: "failed" }
              : item,
          ),
        );
        if (isCurrent())
          toast.error("Message was not sent. Retry it from the conversation.", {
            description: error instanceof Error ? error.message : undefined,
          });
        return undefined;
      }
    },
    [accountId, client, update, workspaceSlug],
  );
  const enqueue = useCallback(
    (conversationId: string, draft: ChatDraft, rootId?: string) => {
      if (!session?.user) return undefined;
      const input = getMessageInput(draft);
      const creationId = crypto.randomUUID();
      const pending: ChatMessage = {
        ...input,
        id: `pending:${creationId}`,
        creationId,
        conversationId,
        threadRootId: rootId ?? null,
        author: {
          id: session.user.id,
          fullName: session.user.name || "You",
          avatarUrl: session.user.image ?? null,
          isSystem: false,
        },
        createdAt: new Date().toISOString(),
        sequence: 0,
        version: 0,
        deliveryStatus: "sending",
        editedAt: null,
        deletedAt: null,
        replyCount: 0,
        unreadReplyCount: 0,
        lastReplyAt: null,
        replyAuthors: [],
        reactions: [],
        attachments: draft.attachments ?? [],
        links: extractMessageLinks(input.text, input.richContent),
        taskRefs: [],
      };
      update((current) => [...current, pending]);
      return { pending, delivery: deliver(pending) };
    },
    [deliver, session?.user, update],
  );
  const retry = useCallback(
    async (messageId: string) => {
      const pending = client
        .getQueryData<ChatMessage[]>(outboxKey(workspaceSlug, accountId))
        ?.find((message) => message.id === messageId);
      if (pending?.deliveryStatus !== "failed") return;
      update((current) =>
        current.map((item) =>
          item.id === messageId ? { ...item, deliveryStatus: "sending" } : item,
        ),
      );
      await deliver(pending);
    },
    [accountId, client, deliver, update, workspaceSlug],
  );
  const acknowledge = useCallback(
    (messages: ChatMessage[]) => {
      const confirmed = new Set(
        messages.map((message) => `${message.author.id}:${message.creationId}`),
      );
      update((current) => {
        const remaining = current.filter(
          (message) =>
            !confirmed.has(`${message.author.id}:${message.creationId}`),
        );
        return remaining.length === current.length ? current : remaining;
      });
    },
    [update],
  );
  return { outgoing, enqueue, retry, acknowledge };
};
