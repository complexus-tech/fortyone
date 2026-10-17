"use client";

import { useCallback, useEffect, useRef } from "react";
import { useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { useSession } from "@/lib/auth/client";
import { useWorkspacePath } from "@/hooks/use-workspace-path";
import type { ChatMessage } from "./types";
import * as api from "./api";
import { recordChatMessageVisible } from "./alerts/active-view";
import { mutateUnreadOptimistically } from "./optimistic-unread";
import { reconcileChatQueries } from "./realtime";

const streamKey = (conversationId: string, threadRootId: string | null) =>
  `${conversationId}:${threadRootId ?? "main"}`;
const isFocused = () =>
  document.visibilityState === "visible" && document.hasFocus();
type ReadView = {
  conversationId: string;
  threadRootId: string | null;
  enabled: boolean;
};
type ReadHold = {
  status: "pending" | "succeeded" | "failed";
  previous?: ReadHold;
};
const isVisibleStream = (view: ReadView, message: ChatMessage) =>
  view.enabled &&
  view.conversationId === message.conversationId &&
  (!message.threadRootId || view.threadRootId === message.threadRootId);

/** Manual unread takes precedence over visibility until the reader leaves this stream. */
export const useChatReadState = ({
  conversationId,
  threadRootId,
  lastReadSequence = 0,
  enabled = true,
}: {
  conversationId: string;
  threadRootId: string | null;
  lastReadSequence?: number;
  enabled?: boolean;
}) => {
  const client = useQueryClient();
  const { workspaceSlug } = useWorkspacePath();
  const { data: session } = useSession();
  const accountId = session?.user.id;
  const scope = `${workspaceSlug.toLowerCase()}:${accountId ?? ""}`;
  const activeScope = useRef<{ scope: string } | null>({ scope });
  const timers = useRef(new Map<string, ReturnType<typeof setTimeout>>());
  const sequences = useRef(new Map<string, number>());
  const paused = useRef(new Map<string, ReadHold>());
  const writes = useRef(new Map<string, Promise<unknown>>());
  const view = useRef<ReadView>({ conversationId, threadRootId, enabled });

  useEffect(() => {
    activeScope.current = { scope };
    const pendingTimers = timers.current;
    const seen = sequences.current;
    const holds = paused.current;
    return () => {
      activeScope.current = null;
      pendingTimers.forEach(clearTimeout);
      pendingTimers.clear();
      seen.clear();
      holds.clear();
    };
  }, [scope]);

  useEffect(() => {
    view.current = { conversationId, threadRootId, enabled };
    const visibleStreams = new Set(
      enabled
        ? [
            streamKey(conversationId, null),
            ...(threadRootId ? [streamKey(conversationId, threadRootId)] : []),
          ]
        : [],
    );
    for (const key of paused.current.keys()) {
      if (!visibleStreams.has(key)) paused.current.delete(key);
    }
    for (const [key, timer] of timers.current) {
      if (!visibleStreams.has(key)) {
        clearTimeout(timer);
        timers.current.delete(key);
        sequences.current.delete(key);
      }
    }
  }, [conversationId, enabled, threadRootId]);

  const recordVisibleMessages = useCallback(
    (messages: readonly ChatMessage[]) => {
      if (!enabled || !isFocused()) return;
      for (const message of messages) {
        if (
          message.deliveryStatus ||
          !isVisibleStream(view.current, message) ||
          paused.current.has(
            streamKey(message.conversationId, message.threadRootId),
          )
        )
          continue;
        recordChatMessageVisible(accountId, workspaceSlug, message);
      }
    },
    [accountId, enabled, workspaceSlug],
  );

  const markVisible = useCallback(
    (message: ChatMessage) => {
      const key = streamKey(message.conversationId, message.threadRootId);
      const epoch = activeScope.current;
      if (
        !enabled ||
        epoch?.scope !== scope ||
        !accountId ||
        message.deliveryStatus ||
        !isVisibleStream(view.current, message) ||
        !isFocused() ||
        paused.current.has(key)
      )
        return;
      recordChatMessageVisible(accountId, workspaceSlug, message);
      const seen =
        sequences.current.get(key) ??
        (message.threadRootId ? 0 : lastReadSequence);
      if (message.sequence <= seen) return;
      sequences.current.set(key, message.sequence);
      clearTimeout(timers.current.get(key));
      timers.current.set(
        key,
        setTimeout(() => {
          timers.current.delete(key);
          const previous = writes.current.get(`${scope}:${key}`);
          // Earlier failures report their own errors; they must not block future reads.
          const write = Promise.resolve(previous)
            .catch(() => undefined)
            .then(async () => {
              if (
                activeScope.current !== epoch ||
                paused.current.has(key) ||
                !isVisibleStream(view.current, message)
              )
                return;
              if (!isFocused()) {
                sequences.current.delete(key);
                return;
              }
              await api.markRead(
                { workspaceSlug },
                message.conversationId,
                message.id,
                message.threadRootId ?? undefined,
              );
              if (activeScope.current === epoch)
                await reconcileChatQueries(
                  client,
                  workspaceSlug,
                  message.conversationId,
                );
            })
            .catch((error: unknown) => {
              if (activeScope.current !== epoch || paused.current.has(key))
                return;
              sequences.current.delete(key);
              toast.error("Could not update the read status", {
                description:
                  error instanceof Error ? error.message : "Please try again.",
              });
            });
          writes.current.set(`${scope}:${key}`, write);
          void write.finally(() => {
            if (writes.current.get(`${scope}:${key}`) === write)
              writes.current.delete(`${scope}:${key}`);
          });
        }, 500),
      );
    },
    [accountId, client, enabled, lastReadSequence, scope, workspaceSlug],
  );

  const markUnread = useCallback(
    async (message: ChatMessage) => {
      const epoch = activeScope.current;
      if (
        epoch?.scope !== scope ||
        !accountId ||
        message.deliveryStatus ||
        message.deletedAt ||
        message.author.id === accountId
      )
        return;
      const key = streamKey(message.conversationId, message.threadRootId);
      const hold: ReadHold = {
        status: "pending",
        previous: paused.current.get(key),
      };
      if (
        enabled &&
        message.conversationId === conversationId &&
        (!message.threadRootId || message.threadRootId === threadRootId)
      )
        paused.current.set(key, hold);
      clearTimeout(timers.current.get(key));
      timers.current.delete(key);
      sequences.current.delete(key);
      const previous = writes.current.get(`${scope}:${key}`);
      const write = mutateUnreadOptimistically(
        client,
        { workspaceSlug, accountId, message },
        async () => {
          // A read already on the wire must finish before the explicit rewind.
          await Promise.resolve(previous).catch(() => undefined);
          if (activeScope.current !== epoch)
            throw new Error("This chat session is no longer active.");
          return api.markUnread(
            { workspaceSlug },
            message.conversationId,
            message.id,
            message.threadRootId ?? undefined,
          );
        },
      )
        .then(() => {
          hold.status = "succeeded";
          hold.previous = undefined;
          if (activeScope.current !== epoch) return;
          toast.success("Marked as unread");
          void reconcileChatQueries(
            client,
            workspaceSlug,
            message.conversationId,
          ).catch((error: unknown) => {
            if (activeScope.current === epoch)
              toast.error(
                "Marked as unread. Could not refresh the conversation.",
                {
                  description:
                    error instanceof Error ? error.message : undefined,
                },
              );
          });
        })
        .catch((error: unknown) => {
          hold.status = "failed";
          if (activeScope.current !== epoch) return;
          if (paused.current.get(key) === hold) {
            let previousHold = hold.previous;
            while (previousHold?.status === "failed")
              previousHold = previousHold.previous;
            if (previousHold) paused.current.set(key, previousHold);
            else paused.current.delete(key);
          }
          toast.error("Could not mark this message as unread", {
            description:
              error instanceof Error ? error.message : "Please try again.",
          });
        });
      writes.current.set(`${scope}:${key}`, write);
      await write;
      if (writes.current.get(`${scope}:${key}`) === write)
        writes.current.delete(`${scope}:${key}`);
    },
    [
      accountId,
      client,
      conversationId,
      enabled,
      scope,
      threadRootId,
      workspaceSlug,
    ],
  );

  return { markUnread, markVisible, recordVisibleMessages };
};
