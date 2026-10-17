"use client";

import { useCallback, useEffect, useMemo, useRef, useState } from "react";
import { useQuery, useQueryClient } from "@tanstack/react-query";
import { useRouter, useSearchParams } from "next/navigation";
import { toast } from "sonner";
import { useWorkspacePath } from "@/hooks/use-workspace-path";
import { useSession } from "@/lib/auth/client";
import type { ChatDraft } from "./chat-rich-text";
import { EMPTY_PREVIEW_DRAFT } from "./chat-rich-text";
import type { ChatMessage, ChatUnreadThread, ConversationTab } from "./types";
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
import { recordChatMessageVisible } from "./alerts/active-view";
import { mergeOutgoingMessages, useMessageOutbox } from "./use-message-outbox";

const NO_MESSAGES: ChatMessage[] = [];
const NEW_MESSAGE_DRAFT_KEY = "new-message";
const notification = (error: unknown, title: string) =>
  toast.error(title, {
    description: error instanceof Error ? error.message : "Please try again.",
  });

export const useChat = () => {
  const { workspaceSlug, withWorkspace } = useWorkspacePath();
  const { data: session } = useSession();
  const accountId = session?.user.id;
  const chatPath = withWorkspace("/chat");
  const router = useRouter();
  const searchParams = useSearchParams();
  const client = useQueryClient();
  const requestedConversation = searchParams.get("conversation") ?? "";
  const requestedNewMessage = searchParams.get("new") === "1";
  const requestedThread = searchParams.get("thread");
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
  const draftsRef = useRef(drafts);
  const updateDrafts = useCallback(
    (
      update: (current: Record<string, ChatDraft>) => Record<string, ChatDraft>,
    ) => {
      draftsRef.current = update(draftsRef.current);
      setDrafts(draftsRef.current);
    },
    [],
  );
  const {
    outgoing,
    enqueue,
    retry: retryMessage,
    acknowledge,
  } = useMessageOutbox();
  const navigationVersion = useRef(0);
  useEffect(
    () => () => {
      navigationVersion.current += 1;
      composing.current = false;
    },
    [],
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
  useEffect(() => {
    navigationVersion.current += 1;
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
  const loadedMessages = useMemo(() => {
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
  const messages = useMemo(
    () =>
      mergeOutgoingMessages(
        loadedMessages,
        outgoing.filter(
          (message) =>
            message.conversationId === activeId && !message.threadRootId,
        ),
      ),
    [activeId, loadedMessages, outgoing],
  );
  const linkedMessage = useChatMessage(requestedMessage);
  const selectedThread = useChatMessage(threadId);
  const thread =
    messages.find((message) => message.id === threadId) ?? selectedThread.data;
  const repliesQuery = useChatMessages(
    threadId ? activeId : "",
    threadTargetId,
    threadId ?? undefined,
  );
  const loadedReplies = useMemo(
    () => flattenMessages(repliesQuery.data?.pages),
    [repliesQuery.data],
  );
  const replies = useMemo(
    () =>
      mergeOutgoingMessages(
        loadedReplies,
        outgoing.filter(
          (message) =>
            message.conversationId === activeId &&
            message.threadRootId === threadId,
        ),
      ),
    [activeId, loadedReplies, outgoing, threadId],
  );
  useEffect(() => {
    acknowledge([...loadedMessages, ...loadedReplies]);
  }, [acknowledge, loadedMessages, loadedReplies]);
  const threadKey = `${activeId}:thread:${threadId ?? ""}`;
  const feed = useMemo(
    () => (conversation ? { ...conversation, messages } : undefined),
    [conversation, messages],
  );
  const hasNewer = Boolean(
    history.hasNextPage ||
      (tail.data?.items.at(-1)?.sequence ?? 0) >
        (loadedMessages.at(-1)?.sequence ?? 0),
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
    setThreadId(
      message.threadRootId ??
        (requestedThread === message.id ? message.id : null),
    );
    setThreadTargetId(message.threadRootId ? message.id : null);
  }, [
    requestedConversation,
    requestedMessage,
    requestedThread,
    requestedNewMessage,
    linkedId,
    linkedRootId,
  ]);
  useEffect(() => {
    if (!linkedMessage.error || !requestedMessage) return;
    notification(linkedMessage.error, "This message is no longer available");
    router.replace(`${chatPath}?conversation=${encodeURIComponent(activeId)}`, {
      scroll: false,
    });
  }, [activeId, chatPath, linkedMessage.error, requestedMessage, router]);
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
      router.replace(`${chatPath}?conversation=${encodeURIComponent(id)}`, {
        scroll: false,
      });
    },
    [chatPath, router],
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
    router.replace(`${chatPath}?new=1`, { scroll: false });
  }, [chatPath, router]);
  const chooseNewMessageRecipient = useCallback(
    (id: string) => {
      if (!composing.current) return;
      const incoming =
        draftsRef.current[NEW_MESSAGE_DRAFT_KEY] ?? EMPTY_PREVIEW_DRAFT;
      updateDrafts((current) => ({
        ...current,
        [id]: mergeChatDrafts(current[id] ?? EMPTY_PREVIEW_DRAFT, incoming),
        [NEW_MESSAGE_DRAFT_KEY]: EMPTY_PREVIEW_DRAFT,
      }));
      selectConversation(id);
    },
    [selectConversation, updateDrafts],
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
    if (requestedThread) {
      const params = new URLSearchParams({ conversation: activeId });
      if (rootId) params.set("message", rootId);
      router.replace(`${chatPath}?${params.toString()}`, { scroll: false });
    }
  }, [activeId, chatPath, messages, requestedThread, router, threadId]);
  const openUnreadThread = useCallback(
    (message: ChatUnreadThread, trigger: HTMLButtonElement) => {
      navigationVersion.current += 1;
      composing.current = false;
      threadTrigger.current = trigger;
      pendingFocus.current = "thread";
      client.setQueryData(
        [...chatKeys.all(workspaceSlug), "message", message.id],
        message,
      );
      setIsComposing(false);
      setSelectedId(message.conversationId);
      setActiveTab("messages");
      setAnchor(message.id);
      setTargetMessageId(message.id);
      setThreadId(message.id);
      setThreadTargetId(message.firstUnreadReplyId);
      setShowConversation(true);
      const params = new URLSearchParams({
        conversation: message.conversationId,
        message: message.firstUnreadReplyId,
        thread: message.id,
      });
      router.replace(`${chatPath}?${params.toString()}`, { scroll: false });
    },
    [chatPath, client, router, workspaceSlug],
  );
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
          `${chatPath}?conversation=${encodeURIComponent(message.conversationId)}&message=${encodeURIComponent(message.id)}`,
          { scroll: false },
        );
      } catch (error) {
        if (requestVersion !== navigationVersion.current) return;
        notification(error, "Could not open this message");
      }
    },
    [chatPath, router, workspaceSlug],
  );
  const updateDraft = useCallback(
    (key: string, draft: ChatDraft) => {
      updateDrafts((current) => ({ ...current, [key]: draft }));
    },
    [updateDrafts],
  );
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
      const draft = draftsRef.current[key] ?? EMPTY_PREVIEW_DRAFT;
      if (
        !activeId ||
        (!draft.text.trim() && !draft.attachments?.length) ||
        draft.text.length > 4000
      )
        return;
      const send = enqueue(activeId, draft, rootId);
      if (!send) return;
      updateDraft(key, EMPTY_PREVIEW_DRAFT);
      setSentMessage(send.pending);
      if (!rootId) {
        setAnchor(null);
        setTargetMessageId(null);
      }
      const message = await send.delivery;
      if (message) {
        setSentMessage((current) =>
          current?.creationId === message.creationId ? message : current,
        );
      }
    },
    [activeId, enqueue, updateDraft],
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
  const recordVisibleMessages = useCallback(
    (visible: readonly ChatMessage[]) => {
      if (document.visibilityState !== "visible" || !document.hasFocus())
        return;
      for (const message of visible) {
        if (message.deliveryStatus) continue;
        recordChatMessageVisible(accountId, workspaceSlug, message);
      }
    },
    [accountId, workspaceSlug],
  );
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
      if (
        message.deliveryStatus ||
        document.visibilityState !== "visible" ||
        !document.hasFocus()
      )
        return;
      if (accountId)
        recordChatMessageVisible(accountId, workspaceSlug, message);
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
              reconcileChatQueries(
                client,
                workspaceSlug,
                message.conversationId,
              ),
            )
            .catch((error: unknown) => {
              readSequences.current.delete(key);
              notification(error, "Could not update the read status");
            });
        }, 500),
      );
    },
    [accountId, client, conversation?.lastReadSequence, workspaceSlug],
  );

  return {
    openUnreadThread,
    recordVisibleMessages,
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
    selectConversation,
    openThread,
    closeThread,
    viewMessage,
    updateDraft,
    sendMessage,
    retryMessage,
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
        router.replace(chatPath, { scroll: false });
      }
      setShowConversation(false);
    },
  };
};
