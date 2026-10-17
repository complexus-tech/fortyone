import type {
  InfiniteData,
  UseInfiniteQueryResult,
} from "@tanstack/react-query";
import { useInfiniteQuery, useQuery } from "@tanstack/react-query";
import { useSession } from "@/lib/auth/client";
import { useWorkspacePath } from "@/hooks/use-workspace-path";
import type {
  ChatMessagePage,
  ChatResourceFilter,
  ChatResourceSort,
} from "./types";
import * as api from "./api";
import { CHAT_MAX_PAGES, chatKeys } from "./constants";

export type MessageCursor = {
  before?: string;
  after?: string;
  around?: string;
};

export const useChatConversations = (search = "") => {
  const { workspaceSlug } = useWorkspacePath();
  const { data: session } = useSession();
  return useInfiniteQuery({
    queryKey: chatKeys.conversations(workspaceSlug, search),
    queryFn: ({ pageParam, signal }) =>
      api.getConversations({ workspaceSlug }, search, pageParam, signal),
    initialPageParam: undefined as string | undefined,
    getNextPageParam: (page) => page.nextCursor ?? undefined,
    enabled: Boolean(session),
    staleTime: 15_000,
  });
};

export const useChatConversation = (conversationId: string) => {
  const { workspaceSlug } = useWorkspacePath();
  return useQuery({
    queryKey: chatKeys.conversation(workspaceSlug, conversationId),
    queryFn: ({ signal }) =>
      api.getConversation({ workspaceSlug }, conversationId, signal),
    enabled: Boolean(conversationId),
    staleTime: 15_000,
  });
};

export const useChatMessages = (
  conversationId: string,
  anchor: string | null,
  threadRootId?: string,
): UseInfiniteQueryResult<InfiniteData<ChatMessagePage, MessageCursor>> => {
  const { workspaceSlug } = useWorkspacePath();
  return useInfiniteQuery({
    queryKey: threadRootId
      ? [
          ...chatKeys.replies(workspaceSlug, conversationId, threadRootId),
          anchor ?? "start",
        ]
      : chatKeys.messages(workspaceSlug, conversationId, anchor ?? "latest"),
    queryFn: ({ pageParam, signal }) =>
      threadRootId
        ? api.getReplies({ workspaceSlug }, threadRootId, pageParam, signal)
        : api.getMessages({ workspaceSlug }, conversationId, pageParam, signal),
    initialPageParam: (anchor ? { around: anchor } : {}) as MessageCursor,
    getPreviousPageParam: (page) =>
      page.hasOlder && page.olderCursor
        ? { before: page.olderCursor }
        : undefined,
    getNextPageParam: (page) =>
      page.hasNewer && page.newerCursor
        ? { after: page.newerCursor }
        : undefined,
    maxPages: CHAT_MAX_PAGES,
    enabled: Boolean(conversationId && (!threadRootId || threadRootId.length)),
    staleTime: 15_000,
  });
};

export const useChatMessage = (messageId: string | null) => {
  const { workspaceSlug } = useWorkspacePath();
  return useQuery({
    queryKey: [...chatKeys.all(workspaceSlug), "message", messageId],
    queryFn: ({ signal }) =>
      api.getMessage({ workspaceSlug }, messageId!, signal),
    enabled: Boolean(messageId),
    staleTime: 15_000,
  });
};

export const useChatSearch = (conversationId: string, search: string) => {
  const { workspaceSlug } = useWorkspacePath();
  return useInfiniteQuery({
    queryKey: chatKeys.search(workspaceSlug, conversationId, search),
    queryFn: ({ pageParam, signal }) =>
      api.searchMessages(
        { workspaceSlug },
        conversationId,
        search,
        pageParam,
        signal,
      ),
    initialPageParam: undefined as string | undefined,
    getNextPageParam: (page) => page.nextCursor ?? undefined,
    maxPages: CHAT_MAX_PAGES,
    enabled: Boolean(conversationId && search.trim()),
    staleTime: 15_000,
  });
};

export const useChatResources = (
  conversationId: string,
  kind: ChatResourceFilter = "all",
  search = "",
  sort: ChatResourceSort = "newest",
  enabled = true,
) => {
  const { workspaceSlug } = useWorkspacePath();
  return useInfiniteQuery({
    queryKey: [
      ...chatKeys.resources(workspaceSlug, conversationId, kind),
      search,
      sort,
    ],
    queryFn: ({ pageParam, signal }) =>
      api.getResources(
        { workspaceSlug },
        conversationId,
        kind,
        search,
        pageParam,
        signal,
        sort,
      ),
    initialPageParam: undefined as string | undefined,
    getNextPageParam: (page) => page.nextCursor ?? undefined,
    // Explicitly loaded resource pages remain visible when loading more.
    // Unlike the message timeline, this list does not compensate for eviction.
    enabled: enabled && Boolean(conversationId),
    staleTime: 15_000,
  });
};
