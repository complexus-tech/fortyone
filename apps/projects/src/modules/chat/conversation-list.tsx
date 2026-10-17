"use client";

import type { ReactNode } from "react";
import { useId, useRef, useState } from "react";
import { ChevronRightIcon, ChatIcon, CloseIcon, SearchIcon } from "icons";
import { cn } from "lib";
import { Button, Flex, Input, Text, Tooltip } from "ui";
import { useDebounce } from "@/hooks/debounce";
import { ConversationMark } from "./avatars";
import { ConversationListActions } from "./conversation-list-actions";
import type { ChatConversation, ChatUnreadThread } from "./types";
import { useChatConversations } from "./queries";
import { useChatUnread } from "./use-chat-unread";
import { ChatUnreadBadge } from "./unread-badge";
import { UnreadThreads } from "./unread-threads";

const ROW_CLASS =
  "hover:bg-surface-muted focus-visible:ring-primary/40 flex w-full items-center gap-2 rounded-lg px-3 py-2 text-left transition-colors outline-none focus-visible:ring-2";

export const ConversationList = ({
  activeId,
  navigationControl,
  onSelect,
  onNewMessage,
  conversations,
  activeThreadId,
  onSelectThread,
}: {
  activeId: string;
  conversations: ChatConversation[];
  navigationControl?: ReactNode;
  onSelect: (id: string) => void;
  onNewMessage: () => void;
  activeThreadId?: string;
  onSelectThread: (
    thread: ChatUnreadThread,
    trigger: HTMLButtonElement,
  ) => void;
}) => {
  const [search, setSearch] = useState("");
  const [query, setQuery] = useState("");
  const searchLater = useDebounce(setQuery, 200);
  const result = useChatConversations(query);
  const unread = useChatUnread();
  const unreadByConversation = new Map(
    unread.data?.conversations.map((conversation) => [
      conversation.conversationId,
      conversation.unreadCount,
    ]),
  );
  const [expanded, setExpanded] = useState({ team: true, dm: true });
  const sectionId = useId();
  const searchRef = useRef<HTMLInputElement>(null);
  const matching = (
    query
      ? result.data?.pages.flatMap((page) => page.items) ?? []
      : conversations
  ).filter((conversation) =>
    conversation.name.toLowerCase().includes(search.trim().toLowerCase()),
  );

  return (
    <>
      <header className="flex h-16 min-w-0 shrink-0 items-center justify-between gap-2 px-4">
        <Flex align="center" className="min-w-0 gap-2">
          {navigationControl}
          <ChatIcon className="h-5 w-auto" />
          <Text as="h1">Chat</Text>
        </Flex>
        <ConversationListActions
          onNewMessage={() => {
            setSearch("");
            setQuery("");
            setExpanded((current) => ({ ...current, dm: true }));
            onNewMessage();
          }}
        />
      </header>
      <search aria-label="Find conversations" className="shrink-0 px-4 pb-1">
        <div className="relative">
          <Input
            aria-controls={`${sectionId}-conversations`}
            aria-label="Search conversations"
            className="h-[calc(2.1rem+2px)] pr-9 text-base [&::-webkit-search-cancel-button]:hidden"
            leftIcon={<SearchIcon className="size-4" />}
            onChange={(event) => {
              setSearch(event.target.value);
              searchLater(event.target.value);
              setExpanded({ team: true, dm: true });
            }}
            onKeyDown={(event) => {
              if (event.key === "Escape") {
                setSearch("");
                setQuery("");
              }
            }}
            placeholder="Search conversations..."
            ref={searchRef}
            size="sm"
            type="search"
            value={search}
          />
          {search ? (
            <Tooltip title="Clear search">
              <Button
                aria-label="Clear conversations search"
                asIcon
                className="absolute top-1/2 right-1 size-6 -translate-y-1/2 rounded-md"
                color="tertiary"
                onClick={() => {
                  setSearch("");
                  setQuery("");
                  searchRef.current?.focus();
                }}
                size="sm"
                type="button"
                variant="naked"
              >
                <CloseIcon className="size-3.5" />
              </Button>
            </Tooltip>
          ) : null}
        </div>
        <span aria-live="polite" className="sr-only" role="status">
          {matching.length}{" "}
          {matching.length === 1 ? "conversation" : "conversations"} found.
        </span>
      </search>
      <nav
        aria-label="Chat conversations"
        className="min-h-0 flex-1 overflow-y-auto px-3 py-3"
        id={`${sectionId}-conversations`}
      >
        <UnreadThreads
          activeThreadId={activeThreadId}
          count={unread.data?.unreadThreads ?? 0}
          onSelect={onSelectThread}
        />
        {(["team", "dm"] as const).map((kind) => (
          <section className="mb-4" key={kind}>
            <h2>
              <button
                aria-controls={`${sectionId}-${kind}`}
                aria-expanded={expanded[kind]}
                className={cn(ROW_CLASS, "text-text-muted font-medium")}
                onClick={() => {
                  setExpanded((current) => ({
                    ...current,
                    [kind]: !current[kind],
                  }));
                }}
                type="button"
              >
                <span className="flex size-6 shrink-0 items-center justify-center">
                  <ChevronRightIcon
                    className={cn(
                      "h-3.5 w-auto text-current transition-transform",
                      expanded[kind] && "rotate-90",
                    )}
                  />
                </span>
                {kind === "team" ? "Teams" : "Direct Messages"}
              </button>
            </h2>
            <div hidden={!expanded[kind]} id={`${sectionId}-${kind}`}>
              {matching
                .filter((conversation) => conversation.kind === kind)
                .map((conversation) => (
                  <button
                    aria-current={
                      activeId === conversation.id ? "page" : undefined
                    }
                    className={cn(
                      ROW_CLASS,
                      "text-foreground pl-8",
                      activeId === conversation.id &&
                        "bg-primary/5 text-primary hover:bg-primary/5 font-medium",
                    )}
                    key={conversation.id}
                    onClick={() => {
                      onSelect(conversation.id);
                    }}
                    type="button"
                  >
                    <ConversationMark
                      active={activeId === conversation.id}
                      conversation={conversation}
                    />
                    <span className="truncate">{conversation.name}</span>
                    <ChatUnreadBadge
                      className="ml-auto"
                      count={
                        unread.data
                          ? unreadByConversation.get(conversation.id) ?? 0
                          : conversation.unreadCount
                      }
                    />
                  </button>
                ))}
            </div>
          </section>
        ))}
        {result.hasNextPage ? (
          <Button
            className="mt-2 w-full"
            color="tertiary"
            disabled={result.isFetchingNextPage}
            onClick={() => {
              void result.fetchNextPage();
            }}
            size="sm"
          >
            Load more conversations
          </Button>
        ) : null}
        {result.error ? (
          <div className="px-3 py-4" role="alert">
            <Text color="muted">Could not load conversations.</Text>
            <Button
              onClick={() => {
                void result.refetch();
              }}
              size="sm"
            >
              Try again
            </Button>
          </div>
        ) : null}
        {matching.length === 0 && !result.error ? (
          <Text className="px-3 py-5" color="muted">
            {result.isFetching
              ? "Loading conversations…"
              : "No conversations found."}
          </Text>
        ) : null}
      </nav>
    </>
  );
};
