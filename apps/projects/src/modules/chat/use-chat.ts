"use client";

import { useCallback, useEffect, useMemo, useRef, useState } from "react";
import { useQuery, useQueryClient } from "@tanstack/react-query";
import { useRouter, useSearchParams } from "next/navigation";
import { toast } from "sonner";
import { useWorkspacePath } from "@/hooks/use-workspace-path";
import type { ChatDraft } from "./chat-rich-text";
import { EMPTY_PREVIEW_DRAFT } from "./chat-rich-text";
import type { ChatMessage, ConversationTab } from "./types";
import * as api from "./api";
import { chatKeys } from "./constants";
import { flattenMessages, updateCachedMessage } from "./cache";
import { getMessageInput } from "./message-input";
import {
  useChatConversation,
  useChatConversations,
  useChatMessage,
  useChatMessages,
} from "./queries";
import { reconcileChatQueries } from "./realtime";
import { mergeChatDrafts } from "./merge-chat-drafts";

const NO_MESSAGES: ChatMessage[] = [];
const NEW_MESSAGE_DRAFT_KEY = "new-message";
const notification = (error: unknown, title: string) =>
  toast.error(title, {
    description: error instanceof Error ? error.message : "Please try again.",
  });

export const useChat = () => {
  const { workspaceSlug } = useWorkspacePath();
  const router = useRouter();
  const searchParams = useSearchParams();
  const client = useQueryClient();
  const requestedConversation = searchParams.get("conversation") ?? "";
  const requestedNewMessage = searchParams.get("new") === "1";
  const requestedMessage = requestedNewMessage
    ? null
    : searchParams.get("message");
  const [isComposing, setIsComposing] = useState(requestedNewMessage);
  const composing = useRef(isComposing);
  composing.current = isComposing;
  const [selectedId, setSelectedId] = useState(requestedConversation);
  const [activeTab, setActiveTab] = useState<ConversationTab>("messages");
  const [showConversation, setShowConversation] = useState(
    requestedNewMessage || Boolean(requestedConversation),
  );
  const [threadId, setThreadId] = useState<string | null>(null);
  const [threadTargetId, setThreadTargetId] = useState<string | null>(null);
  const [targetMessageId, setTargetMessageId] = useState<string | null>(
    requestedMessage,
  );
  const [anchor, setAnchor] = useState<string | null>(requestedMessage);
  const [sentMessage, setSentMessage] = useState<ChatMessage | null>(null);
  const [drafts, setDrafts] = useState<Record<string, ChatDraft>>({});
  const [sendingKeys, setSendingKeys] = useState<ReadonlySet<string>>(
    new Set(),
  );
  const sending = useRef(new Set<string>());
  const draftsAddedDuringSend = useRef(new Map<string, ChatDraft>());
  const navigationVersion = useRef(0);
  useEffect(
    () => () => {
      navigationVersion.current += 1;
      composing.current = false;
    },
    [],
  );
  const identities = useRef(
    new Map<string, { fingerprint: string; id: string }>(),
  );
  const closeThreadRef = useRef<HTMLButtonElement>(null);
  const conversationHeadingRef = useRef<HTMLHeadingElement>(null);
  const conversationListRef = useRef<HTMLElement>(null);
  const threadTrigger = useRef<HTMLButtonElement | null>(null);
  const pendingFocus = useRef<"conversation" | "thread" | "list" | null>(null);
  const conversationList = useChatConversations();
  const conversations = useMemo(
    () => conversationList.data?.pages.flatMap((page) => page.items) ?? [],
    [conversationList.data],
  );
  const activeId = isComposing ? "" : selectedId || conversations[0]?.id || "";
  const currentView = useRef({ conversationId: activeId, threadId });
  useEffect(() => {
    navigationVersion.current += 1;
    currentView.current = { conversationId: activeId, threadId };
  }, [activeId, threadId]);
  const detail = useChatConversation(activeId);
  const conversation =
    detail.data ?? conversations.find((item) => item.id === activeId);
  const history = useChatMessages(activeId, anchor);
  const tail = useQuery({
    queryKey: [...chatKeys.all(workspaceSlug), "latest", activeId],
    queryFn: ({ signal }) =>
      api.getMessages({ workspaceSlug }, activeId, {}, signal),
    enabled: Boolean(activeId),
    staleTime: 15_000,
  });
  const messages = useMemo(() => {
    const pages = history.data?.pages;
    const loaded = flattenMessages(pages);
    const latest = tail.data;
    const canMergeTail = latest?.items.some((item) =>
      loaded.some((existing) => existing.id === item.id),
    );
    return canMergeTail && latest
      ? flattenMessages([...(pages ?? []), latest])
      : loaded;
  }, [history.data, tail.data]);
  const linkedMessage = useChatMessage(requestedMessage);
  const selectedThread = useChatMessage(threadId);
  const thread =
    messages.find((message) => message.id === threadId) ?? selectedThread.data;
  const repliesQuery = useChatMessages(
    threadId ? activeId : "",
    threadTargetId,
    threadId ?? undefined,
  );
  const replies = useMemo(
    () => flattenMessages(repliesQuery.data?.pages),
    [repliesQuery.data],
  );
  const threadKey = `${activeId}:thread:${threadId ?? ""}`;
  const feed = useMemo(
    () => (conversation ? { ...conversation, messages } : undefined),
    [conversation, messages],
  );
  const hasNewer = Boolean(
    history.hasNextPage ||
      (tail.data?.items.at(-1)?.sequence ?? 0) >
        (messages.at(-1)?.sequence ?? 0),
  );

  const linkedId = linkedMessage.data?.id;
  const linkedRootId = linkedMessage.data?.threadRootId;
  useEffect(() => {
    navigationVersion.current += 1;
    setIsComposing(requestedNewMessage);
    if (!requestedNewMessage) setSelectedId(requestedConversation);
    setShowConversation(requestedNewMessage || Boolean(requestedConversation));
    if (!requestedMessage) {
      setAnchor(null);
      setTargetMessageId(null);
      setThreadId(null);
      setThreadTargetId(null);
      return;
    }
    if (!linkedId) return;
    const message = { id: linkedId, threadRootId: linkedRootId ?? null };
    setActiveTab("messages");
    setAnchor(message.threadRootId ?? message.id);
    setTargetMessageId(message.threadRootId ?? message.id);
    setThreadId(message.threadRootId);
    setThreadTargetId(message.threadRootId ? message.id : null);
  }, [
    requestedConversation,
    requestedMessage,
    requestedNewMessage,
    linkedId,
    linkedRootId,
  ]);
  useEffect(() => {
    if (!linkedMessage.error || !requestedMessage) return;
    notification(linkedMessage.error, "This message is no longer available");
    router.replace(
      `/${workspaceSlug}/chat?conversation=${encodeURIComponent(activeId)}`,
      { scroll: false },
    );
  }, [activeId, linkedMessage.error, requestedMessage, router, workspaceSlug]);
  useEffect(() => {
    if (!pendingFocus.current) return;
    if (pendingFocus.current === "thread" && thread)
      closeThreadRef.current?.focus();
    else if (pendingFocus.current === "conversation" && conversation)
      conversationHeadingRef.current?.focus();
    else if (pendingFocus.current === "list")
      conversationListRef.current?.focus();
    else return;
    pendingFocus.current = null;
  }, [conversation, showConversation, thread]);

  const selectConversation = useCallback(
    (id: string) => {
      navigationVersion.current += 1;
      composing.current = false;
      setIsComposing(false);
      setSelectedId(id);
      setActiveTab("messages");
      setTargetMessageId(null);
      setAnchor(null);
      setThreadId(null);
      setThreadTargetId(null);
      setShowConversation(true);
      if (window.matchMedia("(max-width: 1023px)").matches)
        pendingFocus.current = "conversation";
      router.replace(
        `/${workspaceSlug}/chat?conversation=${encodeURIComponent(id)}`,
        { scroll: false },
      );
    },
    [router, workspaceSlug],
  );
  const startNewMessage = useCallback(() => {
    navigationVersion.current += 1;
    composing.current = true;
    setIsComposing(true);
    setActiveTab("messages");
    setShowConversation(true);
    setThreadId(null);
    setThreadTargetId(null);
    setTargetMessageId(null);
    setAnchor(null);
    pendingFocus.current = null;
    router.replace(`/${workspaceSlug}/chat?new=1`, { scroll: false });
  }, [router, workspaceSlug]);
  const chooseNewMessageRecipient = useCallback(
    (id: string) => {
      if (!composing.current) return;
      const incoming = drafts[NEW_MESSAGE_DRAFT_KEY] ?? EMPTY_PREVIEW_DRAFT;
      if (sending.current.has(id)) {
        draftsAddedDuringSend.current.set(
          id,
          mergeChatDrafts(
            draftsAddedDuringSend.current.get(id) ?? EMPTY_PREVIEW_DRAFT,
            incoming,
          ),
        );
      }
      setDrafts((current) => ({
        ...current,
        [id]: mergeChatDrafts(current[id] ?? EMPTY_PREVIEW_DRAFT, incoming),
        [NEW_MESSAGE_DRAFT_KEY]: EMPTY_PREVIEW_DRAFT,
      }));
      selectConversation(id);
    },
    [drafts, selectConversation],
  );
  const openThread = useCallback(
    (message: ChatMessage, trigger: HTMLButtonElement) => {
      threadTrigger.current = trigger;
      pendingFocus.current = "thread";
      setThreadId(message.threadRootId ?? message.id);
      setThreadTargetId(message.threadRootId ? message.id : null);
      setShowConversation(true);
    },
    [],
  );
  const closeThread = useCallback(() => {
    const rootId = threadId;
    setThreadId(null);
    setThreadTargetId(null);
    if (threadTrigger.current?.isConnected) threadTrigger.current.focus();
    else if (rootId) {
      setTargetMessageId(rootId);
      if (!messages.some((message) => message.id === rootId)) setAnchor(rootId);
    }
  }, [messages, threadId]);
  const viewMessage = useCallback(
    async (messageId: string) => {
      const requestVersion = ++navigationVersion.current;
      try {
        const message = await api.getMessage({ workspaceSlug }, messageId);
        if (requestVersion !== navigationVersion.current) return;
        composing.current = false;
        setIsComposing(false);
        setActiveTab("messages");
        setShowConversation(true);
        setAnchor(message.threadRootId ?? message.id);
        setTargetMessageId(message.threadRootId ?? message.id);
        setThreadId(message.threadRootId);
        setThreadTargetId(message.threadRootId ? message.id : null);
        router.replace(
          `/${workspaceSlug}/chat?conversation=${encodeURIComponent(message.conversationId)}&message=${encodeURIComponent(message.id)}`,
          { scroll: false },
        );
      } catch (error) {
        if (requestVersion !== navigationVersion.current) return;
        notification(error, "Could not open this message");
      }
    },
    [router, workspaceSlug],
  );
  const updateDraft = useCallback((key: string, draft: ChatDraft) => {
    setDrafts((current) => ({ ...current, [key]: draft }));
  }, []);
  const loadLatest = useCallback(
    async (conversationId: string) => {
      const latest = await api.getMessages({ workspaceSlug }, conversationId);
      client.setQueryData(
        chatKeys.messages(workspaceSlug, conversationId, "latest"),
        { pages: [latest], pageParams: [{}] },
      );
      client.setQueryData(
        [...chatKeys.all(workspaceSlug), "latest", conversationId],
        latest,
      );
      return latest;
    },
    [client, workspaceSlug],
  );
  const jumpToLatest = useCallback(() => {
    void loadLatest(activeId)
      .then((latest) => {
        setAnchor(null);
        setTargetMessageId(latest.items.at(-1)?.id ?? null);
      })
      .catch((error: unknown) => {
        notification(error, "Could not load the latest messages");
      });
  }, [activeId, loadLatest]);
  const sendMessage = useCallback(
    async (key: string, rootId?: string) => {
      const draft = drafts[key] ?? EMPTY_PREVIEW_DRAFT;
      if (
        !activeId ||
        (!draft.text.trim() && !draft.attachments?.length) ||
        draft.text.length > 4000 ||
        sending.current.has(key)
      )
        return;
      const input = getMessageInput(draft);
      const fingerprint = JSON.stringify({ ...input, rootId });
      const previous = identities.current.get(key);
      const creationId =
        previous?.fingerprint === fingerprint
          ? previous.id
          : crypto.randomUUID();
      identities.current.set(key, { fingerprint, id: creationId });
      sending.current.add(key);
      draftsAddedDuringSend.current.delete(key);
      setSendingKeys(new Set(sending.current));
      let committed = false;
      try {
        const message = await api.createMessage({ workspaceSlug }, activeId, {
          ...input,
          creationId,
          threadRootId: rootId,
        });
        committed = true;
        setSentMessage(message);
        updateCachedMessage(client, workspaceSlug, message);
        const addedDraft = draftsAddedDuringSend.current.get(key);
        setDrafts((current) => ({
          ...current,
          [key]:
            addedDraft ??
            (current[key] === draft ? EMPTY_PREVIEW_DRAFT : current[key]),
        }));
        identities.current.delete(key);
        if (!rootId) {
          await loadLatest(activeId);
          if (currentView.current.conversationId === activeId) {
            setAnchor(null);
            setTargetMessageId(null);
          }
        } else if (
          currentView.current.conversationId === activeId &&
          currentView.current.threadId === rootId
        )
          setThreadTargetId(message.id);
        await reconcileChatQueries(client, workspaceSlug, activeId);
      } catch (error) {
        notification(
          error,
          committed
            ? "Message sent. Could not refresh the conversation."
            : "Message was not sent. Your draft is kept.",
        );
      }
      sending.current.delete(key);
      draftsAddedDuringSend.current.delete(key);
      setSendingKeys(new Set(sending.current));
    },
    [activeId, client, drafts, loadLatest, workspaceSlug],
  );
  const editMessage = useCallback(
    async (messageId: string, draft: ChatDraft) => {
      try {
        const original =
          messages.find((item) => item.id === messageId) ??
          replies.find((item) => item.id === messageId) ??
          (await api.getMessage({ workspaceSlug }, messageId));
        const updated = await api.editMessage({ workspaceSlug }, messageId, {
          ...getMessageInput(draft),
          version: draft.baseVersion ?? original.version,
        });
        updateCachedMessage(client, workspaceSlug, updated);
        await reconcileChatQueries(client, workspaceSlug, activeId);
      } catch (error) {
        notification(error, "Could not save your changes");
        throw error;
      }
    },
    [activeId, client, messages, replies, workspaceSlug],
  );
  const deleteMessage = useCallback(
    async (messageId: string) => {
      try {
        const original =
          messages.find((item) => item.id === messageId) ??
          replies.find((item) => item.id === messageId) ??
          (await api.getMessage({ workspaceSlug }, messageId));
        await api.deleteMessage({ workspaceSlug }, messageId, original.version);
        if (threadId === messageId && !original.replyCount) closeThread();
        await reconcileChatQueries(client, workspaceSlug, activeId);
      } catch (error) {
        notification(error, "Could not delete this message");
        throw error;
      }
    },
    [activeId, client, closeThread, messages, replies, threadId, workspaceSlug],
  );
  const toggleReaction = useCallback(
    async (messageId: string, emoji: string) => {
      try {
        const message =
          messages.find((item) => item.id === messageId) ??
          replies.find((item) => item.id === messageId) ??
          (await api.getMessage({ workspaceSlug }, messageId));
        const selected = !message.reactions.some(
          (reaction) => reaction.emoji === emoji && reaction.reacted,
        );
        const updated = await api.setReaction(
          { workspaceSlug },
          messageId,
          emoji,
          selected,
        );
        updateCachedMessage(client, workspaceSlug, updated);
        await reconcileChatQueries(client, workspaceSlug, activeId);
      } catch (error) {
        notification(error, "Could not update this reaction");
      }
    },
    [activeId, client, messages, replies, workspaceSlug],
  );

  const readTimers = useRef(new Map<string, ReturnType<typeof setTimeout>>());
  const readSequences = useRef(new Map<string, number>());
  useEffect(() => {
    const timers = readTimers.current;
    return () => {
      timers.forEach(clearTimeout);
      timers.clear();
    };
  }, []);
  const markVisible = useCallback(
    (message: ChatMessage) => {
      if (document.visibilityState !== "visible" || !document.hasFocus())
        return;
      const key = `${message.conversationId}:${message.threadRootId ?? "root"}`;
      const seen =
        readSequences.current.get(key) ??
        (message.threadRootId ? 0 : conversation?.lastReadSequence ?? 0);
      if (message.sequence <= seen) return;
      readSequences.current.set(key, message.sequence);
      clearTimeout(readTimers.current.get(key));
      readTimers.current.set(
        key,
        setTimeout(() => {
          readTimers.current.delete(key);
          if (document.visibilityState !== "visible" || !document.hasFocus()) {
            readSequences.current.delete(key);
            return;
          }
          void api
            .markRead(
              { workspaceSlug },
              message.conversationId,
              message.id,
              message.threadRootId ?? undefined,
            )
            .then(() =>
              client.invalidateQueries({
                queryKey: chatKeys.conversations(workspaceSlug),
              }),
            )
            .catch((error: unknown) => {
              readSequences.current.delete(key);
              notification(error, "Could not update the read status");
            });
        }, 500),
      );
    },
    [client, conversation?.lastReadSequence, workspaceSlug],
  );

  return {
    isComposing,
    startNewMessage,
    chooseNewMessageRecipient,
    newMessageDraft: drafts[NEW_MESSAGE_DRAFT_KEY] ?? EMPTY_PREVIEW_DRAFT,
    updateNewMessageDraft: (draft: ChatDraft) => {
      updateDraft(NEW_MESSAGE_DRAFT_KEY, draft);
    },
    activeId,
    activeTab,
    conversation: feed,
    conversations,
    conversationList,
    detail,
    history,
    hasNewer,
    showConversation,
    thread,
    threadId,
    threadKey,
    threadTargetId,
    targetMessageId,
    sentMessage,
    replies: threadId ? replies : NO_MESSAGES,
    repliesQuery,
    linkedMessage,
    closeThreadRef,
    conversationHeadingRef,
    conversationListRef,
    draft: drafts[activeId] ?? EMPTY_PREVIEW_DRAFT,
    threadDraft: drafts[threadKey] ?? EMPTY_PREVIEW_DRAFT,
    sendingKeys,
    selectConversation,
    openThread,
    closeThread,
    viewMessage,
    updateDraft,
    sendMessage,
    editMessage,
    deleteMessage,
    toggleReaction,
    markVisible,
    jumpToLatest,
    selectTab: (tab: ConversationTab) => {
      setActiveTab(tab);
      setThreadId(null);
    },
    showList: () => {
      navigationVersion.current += 1;
      pendingFocus.current = "list";
      if (composing.current) {
        composing.current = false;
        setIsComposing(false);
        router.replace(`/${workspaceSlug}/chat`, { scroll: false });
      }
      setShowConversation(false);
    },
  };
};
