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
import { flattenMessages, getCachedMessage } from "./cache";
import { getMessageInput } from "./message-input";
import {
  useChatConversation,
  useChatConversations,
  useChatMessage,
  useChatMessages,
} from "./queries";
import { reconcileChatQueries } from "./realtime";
import { mergeChatDrafts } from "./merge-chat-drafts";
import { useChatReadState } from "./use-chat-read-state";
import { mergeOutgoingMessages, useMessageOutbox } from "./use-message-outbox";
import { withOptimisticReplySummary } from "./optimistic-reply-summaries";
import { extractMessageLinks } from "./chat-links";
import {
  commitMessageReaction,
  deletedMessage,
  mutateMessageOptimistically,
  setMessageReaction,
} from "./optimistic-messages";

const NO_MESSAGES: ChatMessage[] = [];
const NEW_MESSAGE_DRAFT_KEY = "new-message";
const notification = (error: unknown, title: string) =>
  toast.error(title, {
    description: error instanceof Error ? error.message : "Please try again.",
  });
const refreshAfterMutation = (
  client: ReturnType<typeof useQueryClient>,
  workspaceSlug: string,
  conversationId: string,
) => {
  void reconcileChatQueries(client, workspaceSlug, conversationId).catch(
    (error: unknown) =>
      notification(error, "Changes saved. Could not refresh the conversation."),
  );
};

export const useChat = () => {
  const { workspaceSlug, withWorkspace } = useWorkspacePath();
  const { data: session } = useSession();
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
      ).map((message) => withOptimisticReplySummary(message, outgoing)),
    [activeId, loadedMessages, outgoing],
  );
  const linkedMessage = useChatMessage(requestedMessage);
  const selectedThread = useChatMessage(threadId);
  const thread =
    messages.find((message) => message.id === threadId) ??
    (selectedThread.data
      ? withOptimisticReplySummary(selectedThread.data, outgoing)
      : undefined);
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
          getCachedMessage(client, workspaceSlug, messageId) ??
          messages.find((item) => item.id === messageId) ??
          replies.find((item) => item.id === messageId) ??
          (await api.getMessage({ workspaceSlug }, messageId));
        const input = getMessageInput(draft);
        const editedAt = new Date().toISOString();
        await mutateMessageOptimistically(client, workspaceSlug, original, {
          apply: (current) =>
            current.deletedAt
              ? current
              : {
                  ...current,
                  ...input,
                  editedAt,
                  attachments: draft.attachments ?? [],
                  links: extractMessageLinks(input.text, input.richContent),
                },
          execute: () =>
            api.editMessage({ workspaceSlug }, messageId, {
              ...input,
              version: draft.baseVersion ?? original.version,
            }),
          commit: (current, response) => ({
            ...current,
            ...input,
            text: response.text,
            richContent: response.richContent,
            editedAt: response.editedAt,
            version: response.version,
            attachments: response.attachments,
            links: response.links,
          }),
        });
        refreshAfterMutation(client, workspaceSlug, original.conversationId);
      } catch (error) {
        notification(error, "Could not save your changes");
        throw error;
      }
    },
    [client, messages, replies, workspaceSlug],
  );
  const deleteMessage = useCallback(
    async (messageId: string) => {
      try {
        const original =
          getCachedMessage(client, workspaceSlug, messageId) ??
          messages.find((item) => item.id === messageId) ??
          replies.find((item) => item.id === messageId) ??
          (await api.getMessage({ workspaceSlug }, messageId));
        const deletedAt = new Date().toISOString();
        await mutateMessageOptimistically(client, workspaceSlug, original, {
          apply: (current) => deletedMessage(current, deletedAt),
          execute: async (current) => {
            const response = await api.deleteMessage(
              { workspaceSlug },
              messageId,
              current.version,
            );
            return {
              ...deletedMessage(current, response.deletedAt),
              version: response.version,
            };
          },
          commit: (current, response) => ({
            ...deletedMessage(current, response.deletedAt!),
            version: response.version,
          }),
        });
        if (threadId === messageId && !original.replyCount) closeThread();
        refreshAfterMutation(client, workspaceSlug, original.conversationId);
      } catch (error) {
        notification(error, "Could not delete this message");
        throw error;
      }
    },
    [client, closeThread, messages, replies, threadId, workspaceSlug],
  );
  const toggleReaction = useCallback(
    async (messageId: string, emoji: string) => {
      try {
        const message =
          getCachedMessage(client, workspaceSlug, messageId) ??
          messages.find((item) => item.id === messageId) ??
          replies.find((item) => item.id === messageId) ??
          (await api.getMessage({ workspaceSlug }, messageId));
        const selected = !message.reactions.some(
          (reaction) => reaction.emoji === emoji && reaction.reacted,
        );
        if (!session?.user || message.deletedAt) return;
        const person = {
          id: session.user.id,
          fullName: session.user.name || "You",
          avatarUrl: session.user.image ?? null,
          isSystem: false,
        };
        await mutateMessageOptimistically(client, workspaceSlug, message, {
          apply: (current) =>
            setMessageReaction(current, emoji, selected, person),
          execute: () =>
            api.setReaction({ workspaceSlug }, messageId, emoji, selected),
          commit: (current, response, refreshed) =>
            commitMessageReaction(
              current,
              response,
              emoji,
              selected,
              person,
              refreshed,
            ),
        });
        refreshAfterMutation(client, workspaceSlug, message.conversationId);
      } catch (error) {
        notification(error, "Could not update this reaction");
      }
    },
    [client, messages, replies, session, workspaceSlug],
  );

  const { markUnread, markVisible, recordVisibleMessages } = useChatReadState({
    conversationId: activeId,
    threadRootId: threadId,
    lastReadSequence: conversation?.lastReadSequence,
    enabled: !isComposing && activeTab === "messages",
  });

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
    markUnread,
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
