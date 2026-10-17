"use client";

import { useId, useState } from "react";
import { ChevronRightIcon } from "icons";
import { cn } from "lib";
import { Button, Text } from "ui";
import { useUnreadThreads } from "./queries";
import { ChatUnreadBadge } from "./unread-badge";
import type { ChatUnreadThread } from "./types";

export const UnreadThreads = ({
  count,
  activeThreadId,
  onSelect,
}: {
  count: number;
  activeThreadId?: string;
  onSelect: (thread: ChatUnreadThread, trigger: HTMLButtonElement) => void;
}) => {
  const [expanded, setExpanded] = useState(false);
  const listId = useId();
  const query = useUnreadThreads(expanded);
  const threads = query.data?.pages.flatMap((page) => page.items) ?? [];

  return (
    <section className="mb-4">
      <h2>
        <button
          aria-controls={listId}
          aria-expanded={expanded}
          className="hover:bg-surface-muted focus-visible:ring-primary/40 flex w-full items-center gap-2 rounded-lg px-3 py-2 text-left font-medium outline-none focus-visible:ring-2"
          onClick={() => {
            setExpanded((value) => !value);
          }}
          type="button"
        >
          <span className="flex size-6 shrink-0 items-center justify-center">
            <ChevronRightIcon
              className={cn(
                "h-3.5 w-auto transition-transform",
                expanded && "rotate-90",
              )}
            />
          </span>
          Threads
          <ChatUnreadBadge
            className="ml-auto"
            count={count}
            label="unread threads"
          />
        </button>
      </h2>
      <div hidden={!expanded} id={listId}>
        {threads.map((thread) => (
          <button
            aria-current={activeThreadId === thread.id ? "page" : undefined}
            className={cn(
              "hover:bg-surface-muted focus-visible:ring-primary/40 mx-1 flex w-[calc(100%-0.5rem)] flex-col gap-1 rounded-lg py-3 pr-3 pl-10 text-left outline-none focus-visible:ring-2",
              activeThreadId === thread.id && "bg-primary/5",
            )}
            key={thread.id}
            onClick={(event) => {
              onSelect(thread, event.currentTarget);
            }}
            type="button"
          >
            <span className="flex w-full items-center gap-2">
              <span className="text-text-muted truncate text-xs">
                {thread.conversationName}
              </span>
              <ChatUnreadBadge
                className="ml-auto"
                count={thread.unreadReplyCount}
                label="unread replies"
              />
            </span>
            <span className="line-clamp-2 text-sm leading-5">
              {thread.deletedAt
                ? "Deleted message"
                : thread.text || "Shared an attachment"}
            </span>
            <span className="text-text-muted truncate text-xs">
              {thread.author.fullName}
            </span>
          </button>
        ))}
        {query.isPending && expanded ? (
          <Text className="px-4 py-3" color="muted" role="status">
            Loading unread threads…
          </Text>
        ) : null}
        {query.error ? (
          <div className="px-4 py-3" role="alert">
            <Text color="muted">Could not load unread threads.</Text>
            <Button onClick={() => void query.refetch()} size="sm">
              Try again
            </Button>
          </div>
        ) : null}
        {!query.isPending && !query.error && threads.length === 0 ? (
          <Text className="px-4 py-3" color="muted">
            You’re all caught up on threads.
          </Text>
        ) : null}
        {query.hasNextPage ? (
          <Button
            className="mt-2 w-full"
            color="tertiary"
            disabled={query.isFetchingNextPage}
            onClick={() => void query.fetchNextPage()}
            size="sm"
          >
            {query.isFetchingNextPage
              ? "Loading threads…"
              : "Load more threads"}
          </Button>
        ) : null}
      </div>
    </section>
  );
};
