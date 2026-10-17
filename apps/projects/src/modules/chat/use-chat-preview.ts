"use client";

import { useCallback, useEffect, useMemo, useRef, useState } from "react";
import type { PreviewDraft } from "./chat-rich-text";
import { EMPTY_PREVIEW_DRAFT } from "./chat-rich-text";
import type { PreviewMessage, PreviewReaction } from "./preview-data";
import {
  PREVIEW_CONVERSATIONS,
  PREVIEW_TODAY,
  PREVIEW_YOU,
} from "./preview-data";

export type ConversationTab = "messages" | "files";

type FocusTarget =
  | { kind: "conversation" | "list" | "thread" }
  | { kind: "thread-return"; messageId: string };

const isCompactViewport = () =>
  window.matchMedia("(max-width: 1023px)").matches;

const getReactions = (message: PreviewMessage, selected: readonly string[]) => {
  const reactions = new Map<string, PreviewReaction>(
    message.reactions?.map((reaction) => [reaction.emoji, { ...reaction }]),
  );
  for (const emoji of selected) {
    const existing = reactions.get(emoji);
    reactions.set(emoji, {
      emoji,
      count: (existing?.count ?? 0) + 1,
      reacted: true,
    });
  }
  return Array.from(reactions.values());
};

// All interaction state is session-local until the chat UI is approved.
export const useChatPreview = () => {
  const [activeId, setActiveId] = useState(PREVIEW_CONVERSATIONS[0].id);
  const [activeTab, setActiveTab] = useState<ConversationTab>("messages");
  const [showConversation, setShowConversation] = useState(false);
  const [threadId, setThreadId] = useState<string | null>(null);
  const [targetMessageId, setTargetMessageId] = useState<string | null>(null);
  const [reactedMessages, setReactedMessages] = useState<
    Partial<Record<string, readonly string[]>>
  >({});
  const threadTriggerRef = useRef<HTMLButtonElement | null>(null);
  const closeThreadRef = useRef<HTMLButtonElement>(null);
  const conversationHeadingRef = useRef<HTMLHeadingElement>(null);
  const conversationListRef = useRef<HTMLElement>(null);
  const pendingFocusRef = useRef<FocusTarget | null>(null);
  const [drafts, setDrafts] = useState<Partial<Record<string, PreviewDraft>>>(
    {},
  );
  const [addedMessages, setAddedMessages] = useState<
    Partial<Record<string, PreviewMessage[]>>
  >({});
  const [editedMessages, setEditedMessages] = useState<
    Partial<Record<string, PreviewDraft>>
  >({});
  const [deletedMessages, setDeletedMessages] = useState<ReadonlySet<string>>(
    new Set(),
  );
  const conversations = useMemo(() => {
    const prepareMessage = (message: PreviewMessage): PreviewMessage => {
      const edit = editedMessages[message.id];
      return {
        ...message,
        ...(edit
          ? { text: edit.text, richContent: edit.content, edited: true }
          : {}),
        reactions: getReactions(message, reactedMessages[message.id] ?? []),
      };
    };
    return PREVIEW_CONVERSATIONS.map((item) => ({
      ...item,
      messages: [...item.messages, ...(addedMessages[item.id] ?? [])].flatMap(
        (message) => {
          const replies = [
            ...(message.replies ?? []),
            ...(addedMessages[`${item.id}:thread:${message.id}`] ?? []),
          ]
            .filter((reply) => !deletedMessages.has(reply.id))
            .map(prepareMessage);
          if (deletedMessages.has(message.id)) {
            return replies.length > 0
              ? [
                  {
                    ...message,
                    text: "",
                    richContent: null,
                    tasks: undefined,
                    reactions: [],
                    deleted: true,
                    replies,
                  },
                ]
              : [];
          }
          return [{ ...prepareMessage(message), replies }];
        },
      ),
    }));
  }, [addedMessages, deletedMessages, editedMessages, reactedMessages]);
  const conversation =
    conversations.find((item) => item.id === activeId) ?? conversations[0];
  const thread = useMemo(
    () =>
      threadId
        ? conversation.messages.find((message) => message.id === threadId)
        : undefined,
    [conversation.messages, threadId],
  );
  const threadKey = thread ? `${activeId}:thread:${thread.id}` : "";

  useEffect(() => {
    const target = pendingFocusRef.current;
    if (!target) return;
    pendingFocusRef.current = null;

    switch (target.kind) {
      case "conversation":
        conversationHeadingRef.current?.focus();
        break;
      case "list":
        conversationListRef.current?.focus();
        break;
      case "thread":
        closeThreadRef.current?.focus();
        break;
      case "thread-return": {
        const trigger = threadTriggerRef.current;
        const message = document.getElementById(
          `chat-message-${target.messageId}`,
        );
        const fallback = message ?? conversationHeadingRef.current;
        (trigger?.isConnected ? trigger : fallback)?.focus();
        break;
      }
    }
  }, [activeId, deletedMessages, showConversation, threadId]);

  const clearThread = useCallback(() => {
    pendingFocusRef.current = null;
    threadTriggerRef.current = null;
    setThreadId(null);
  }, []);
  const selectConversation = useCallback(
    (id: string) => {
      setActiveId(id);
      setActiveTab("messages");
      setTargetMessageId(null);
      clearThread();
      if (isCompactViewport())
        pendingFocusRef.current = { kind: "conversation" };
      setShowConversation(true);
    },
    [clearThread],
  );
  const openThread = useCallback(
    (message: PreviewMessage, trigger: HTMLButtonElement) => {
      threadTriggerRef.current = trigger;
      if (threadId === message.id) {
        closeThreadRef.current?.focus();
        return;
      }
      pendingFocusRef.current = { kind: "thread" };
      setThreadId(message.id);
      setShowConversation(true);
    },
    [threadId],
  );
  const selectTab = useCallback(
    (tab: ConversationTab) => {
      setActiveTab(tab);
      clearThread();
    },
    [clearThread],
  );
  const viewMessage = useCallback((messageId: string) => {
    setActiveTab("messages");
    setTargetMessageId(messageId);
  }, []);
  const updateDraft = useCallback((key: string, value: PreviewDraft) => {
    setDrafts((current) => ({ ...current, [key]: value }));
  }, []);
  const toggleReaction = useCallback((messageId: string, emoji: string) => {
    setReactedMessages((current) => {
      const selected = current[messageId] ?? [];
      return {
        ...current,
        [messageId]: selected.includes(emoji)
          ? selected.filter((item) => item !== emoji)
          : [...selected, emoji],
      };
    });
  }, []);
  const ownMessageIds = useMemo(
    () =>
      new Set(
        conversation.messages
          .flatMap((message) => [message, ...message.replies])
          .filter(
            (message) =>
              message.author.id === PREVIEW_YOU.id && !message.deleted,
          )
          .map((message) => message.id),
      ),
    [conversation.messages],
  );
  const editMessage = useCallback(
    (messageId: string, draft: PreviewDraft) => {
      if (!draft.text.trim() || !ownMessageIds.has(messageId)) return;
      setEditedMessages((current) => ({ ...current, [messageId]: draft }));
    },
    [ownMessageIds],
  );
  const deleteMessage = useCallback(
    (messageId: string) => {
      if (!ownMessageIds.has(messageId)) return;
      const closesThread =
        thread &&
        ((thread.id === messageId && thread.replies.length === 0) ||
          (deletedMessages.has(thread.id) &&
            thread.replies.every((reply) => reply.id === messageId)));
      if (closesThread) {
        setThreadId(null);
        threadTriggerRef.current = null;
      }
      pendingFocusRef.current = {
        kind: thread && !closesThread ? "thread" : "conversation",
      };
      setTargetMessageId((current) => (current === messageId ? null : current));
      setDeletedMessages((current) => new Set(current).add(messageId));
    },
    [deletedMessages, ownMessageIds, thread],
  );
  const sendPreview = useCallback(
    (key: string) => {
      const draft = drafts[key];
      const text = draft?.text.trim();
      if (!text) return;
      const message: PreviewMessage = {
        id: crypto.randomUUID(),
        author: PREVIEW_YOU,
        time: "Just now",
        date: PREVIEW_TODAY,
        text,
        richContent: draft?.content,
      };
      setAddedMessages((current) => ({
        ...current,
        [key]: [...(current[key] ?? []), message],
      }));
      updateDraft(key, EMPTY_PREVIEW_DRAFT);
    },
    [drafts, updateDraft],
  );
  const closeThread = useCallback(() => {
    if (!threadId) return;
    pendingFocusRef.current = { kind: "thread-return", messageId: threadId };
    if (
      !threadTriggerRef.current?.isConnected &&
      !document.getElementById(`chat-message-${threadId}`)
    ) {
      setTargetMessageId(threadId);
    }
    setThreadId(null);
  }, [threadId]);
  const showList = useCallback(() => {
    pendingFocusRef.current = { kind: "list" };
    setShowConversation(false);
  }, []);

  return {
    activeId,
    activeTab,
    showConversation,
    targetMessageId,
    closeThreadRef,
    conversationHeadingRef,
    conversationListRef,
    conversation,
    thread,
    threadKey,
    draft: drafts[activeId] ?? EMPTY_PREVIEW_DRAFT,
    threadDraft: drafts[threadKey] ?? EMPTY_PREVIEW_DRAFT,
    selectConversation,
    openThread,
    selectTab,
    viewMessage,
    toggleReaction,
    editMessage,
    deleteMessage,
    updateDraft,
    sendPreview,
    closeThread,
    showList,
  };
};
