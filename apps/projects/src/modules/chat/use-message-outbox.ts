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
import { getCachedMessage, updateCachedMessage } from "./cache";
import { reconcileChatQueries } from "./realtime";
import type { OutgoingChatMessage } from "./optimistic-reply-summaries";

const EMPTY_OUTBOX: OutgoingChatMessage[] = [];
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
  const { data: outgoing = EMPTY_OUTBOX } = useQuery<OutgoingChatMessage[]>({
    queryKey: outboxKey(workspaceSlug, accountId),
    queryFn: skipToken,
    initialData: EMPTY_OUTBOX,
    enabled: false,
    gcTime: Infinity,
  });
  const update = useCallback(
    (transform: (messages: OutgoingChatMessage[]) => OutgoingChatMessage[]) => {
      client.setQueryData<OutgoingChatMessage[]>(
        outboxKey(workspaceSlug, accountId),
        (current = EMPTY_OUTBOX) => transform(current),
      );
    },
    [accountId, client, workspaceSlug],
  );
  const deliver = useCallback(
    async (pending: OutgoingChatMessage) => {
      const scope = activeScope.current;
      const isCurrent = () => scope !== null && activeScope.current === scope;
      const refreshReplySummary = () => {
        if (!pending.threadRootId || !pending.optimisticReplySummary) return;
        const belongsToRoot = (item: OutgoingChatMessage) =>
          item.conversationId === pending.conversationId &&
          item.threadRootId === pending.threadRootId &&
          Boolean(item.optimisticReplySummary);
        const cohort = (
          client.getQueryData<OutgoingChatMessage[]>(
            outboxKey(workspaceSlug, accountId),
          ) ?? []
        ).filter(belongsToRoot);
        if (
          !cohort.length ||
          cohort.some((item) => !item.optimisticReplySummary?.settled)
        )
          return;
        const isSameCohort = (current: OutgoingChatMessage[]) => {
          const remaining = current.filter(belongsToRoot);
          // A new send or retry joined while the root read was in flight.
          // Keep the confirmed prefix until that whole cohort is reconciled.
          return (
            remaining.length === cohort.length &&
            remaining.every((item) =>
              cohort.some(
                (original) =>
                  original.creationId === item.creationId &&
                  original.optimisticReplySummary ===
                    item.optimisticReplySummary,
              ),
            )
          );
        };
        const currentCohort = () =>
          isSameCohort(
            client.getQueryData<OutgoingChatMessage[]>(
              outboxKey(workspaceSlug, accountId),
            ) ?? [],
          );
        const clearReplySummary = () => {
          update((current) => {
            if (!isSameCohort(current)) return current;
            return current.flatMap((item) => {
              if (!belongsToRoot(item)) return [item];
              return item.acknowledged
                ? []
                : [{ ...item, optimisticReplySummary: undefined }];
            });
          });
        };
        if (!isCurrent()) {
          clearReplySummary();
          return;
        }
        if (cohort.every((item) => item.deliveryStatus === "failed")) {
          clearReplySummary();
          return;
        }
        // A successful send does not return its parent's summary. Fetch that
        // authorized snapshot after the write, then retire only this overlay.
        void api
          .getMessage({ workspaceSlug }, pending.threadRootId)
          .then((root) => {
            if (isCurrent() && currentCohort())
              updateCachedMessage(client, workspaceSlug, root);
          })
          .catch((error: unknown) => {
            if (isCurrent() && currentCohort())
              toast.error("Reply sent. Could not refresh the thread.", {
                description: error instanceof Error ? error.message : undefined,
              });
          })
          .finally(clearReplySummary);
      };
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
            item.creationId === pending.creationId
              ? {
                  ...message,
                  optimisticReplySummary: item.optimisticReplySummary
                    ? { ...item.optimisticReplySummary, settled: true }
                    : undefined,
                  acknowledged: item.acknowledged,
                }
              : item,
          ),
        );
        // The account-scoped outbox survives navigation, but a late response
        // must not populate shared caches or interrupt a different session.
        if (!isCurrent()) {
          refreshReplySummary();
          return undefined;
        }
        updateCachedMessage(client, workspaceSlug, message);
        refreshReplySummary();
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
          .getQueryData<
            OutgoingChatMessage[]
          >(outboxKey(workspaceSlug, accountId))
          ?.some(
            (message) =>
              message.creationId === pending.creationId &&
              !message.acknowledged,
          );
        if (!stillPending) {
          update((current) =>
            current.map((item) =>
              item.creationId === pending.creationId &&
              item.optimisticReplySummary
                ? {
                    ...item,
                    optimisticReplySummary: {
                      ...item.optimisticReplySummary,
                      settled: true,
                    },
                  }
                : item,
            ),
          );
          refreshReplySummary();
          return undefined;
        }
        update((current) =>
          current.map((item) =>
            item.creationId === pending.creationId
              ? {
                  ...item,
                  deliveryStatus: "failed",
                  optimisticReplySummary: item.optimisticReplySummary
                    ? { ...item.optimisticReplySummary, settled: true }
                    : undefined,
                }
              : item,
          ),
        );
        refreshReplySummary();
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
      const root = rootId
        ? getCachedMessage(client, workspaceSlug, rootId)
        : undefined;
      const pending: OutgoingChatMessage = {
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
        optimisticReplySummary: root
          ? { baseReplyCount: root.replyCount }
          : undefined,
      };
      update((current) => [...current, pending]);
      return { pending, delivery: deliver(pending) };
    },
    [client, deliver, session?.user, update, workspaceSlug],
  );
  const retry = useCallback(
    async (messageId: string) => {
      const pending = client
        .getQueryData<
          OutgoingChatMessage[]
        >(outboxKey(workspaceSlug, accountId))
        ?.find((message) => message.id === messageId);
      if (pending?.deliveryStatus !== "failed") return;
      const root = pending.threadRootId
        ? getCachedMessage(client, workspaceSlug, pending.threadRootId)
        : undefined;
      const retrying: OutgoingChatMessage = {
        ...pending,
        deliveryStatus: "sending",
        optimisticReplySummary: root
          ? { baseReplyCount: root.replyCount }
          : undefined,
      };
      update((current) => [
        ...current.filter((item) => item.id !== messageId),
        retrying,
      ]);
      await deliver(retrying);
    },
    [accountId, client, deliver, update, workspaceSlug],
  );
  const acknowledge = useCallback(
    (messages: ChatMessage[]) => {
      const confirmed = new Map(
        messages.map((message) => [
          `${message.author.id}:${message.creationId}`,
          message,
        ]),
      );
      update((current) => {
        const remaining = current.flatMap((message) => {
          const delivered = confirmed.get(
            `${message.author.id}:${message.creationId}`,
          );
          if (!delivered || message.acknowledged) return [message];
          return message.optimisticReplySummary
            ? [
                {
                  ...delivered,
                  acknowledged: true,
                  optimisticReplySummary: message.optimisticReplySummary,
                },
              ]
            : [];
        });
        return remaining.length === current.length &&
          remaining.every((item, index) => item === current[index])
          ? current
          : remaining;
      });
    },
    [update],
  );
  return { outgoing, enqueue, retry, acknowledge };
};
