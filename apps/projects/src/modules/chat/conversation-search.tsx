"use client";

import type { ReactNode } from "react";
import { useRef } from "react";
import { CloseIcon, ReplyIcon } from "icons";
import { Button, Text } from "ui";
import { ExpandableSearchHeader } from "@/components/shared/expandable-search-header";
import type { PreviewDraft } from "./chat-rich-text";
import type { MessageReplyHandler } from "./message-row";
import { MessageRow } from "./message-row";
import type { ChatConversation, ChatMessage } from "./types";
import { useChatSearch } from "./queries";
import { Timeline } from "./timeline";

export const ConversationSearchHeader = ({
  leading,
  query,
  onActivate,
  onSearch,
}: {
  leading: ReactNode;
  query: string;
  onActivate: () => void;
  onSearch: (value: string) => void;
}) => {
  const headerRef = useRef<HTMLDivElement>(null);
  const restoreSearchButton = () => {
    requestAnimationFrame(() => {
      headerRef.current
        ?.querySelector<HTMLButtonElement>(
          'button[aria-label="Search this conversation"]',
        )
        ?.focus();
    });
  };

  return (
    <div
      className="shrink-0"
      onFocusCapture={(event) => {
        if (event.target instanceof HTMLInputElement) onActivate();
      }}
      onKeyDownCapture={(event) => {
        if (event.key === "Escape") restoreSearchButton();
      }}
      ref={headerRef}
    >
      <ExpandableSearchHeader
        actions={
          query ? (
            <Button
              aria-label="Clear conversation search"
              asIcon
              color="tertiary"
              onClick={() => {
                onSearch("");
                restoreSearchButton();
              }}
              size="sm"
              type="button"
            >
              <CloseIcon className="h-4 w-auto" />
            </Button>
          ) : undefined
        }
        className="border-border"
        initialValue={query}
        key={query}
        label="Search this conversation"
        leading={leading}
        onSubmit={(value) => {
          onSearch(value);
          if (!value) {
            restoreSearchButton();
            return;
          }
          requestAnimationFrame(() => {
            headerRef.current
              ?.querySelector<HTMLInputElement>("input")
              ?.focus();
          });
        }}
        placeholder="Search this conversation..."
      />
    </div>
  );
};

export const ConversationSearchResults = ({
  conversation,
  query,
  mayaAvatarUrl,
  onViewMessage,
  onReply,
  onOpenResources,
  onToggleReaction,
  onEditMessage,
  onDeleteMessage,
  onMarkUnread,
}: {
  conversation: ChatConversation;
  query: string;
  mayaAvatarUrl?: string | null;
  onViewMessage: (messageId: string) => void;
  onReply: MessageReplyHandler;
  onOpenResources: () => void;
  onToggleReaction: (messageId: string, emoji: string) => void;
  onEditMessage: (
    messageId: string,
    draft: PreviewDraft,
  ) => void | Promise<void>;
  onDeleteMessage: (messageId: string) => void | Promise<void>;
  onMarkUnread?: (message: ChatMessage) => void;
}) => {
  const results = useChatSearch(conversation.id, query);
  const matches = results.data?.pages.flatMap((page) => page.items) ?? [];
  return (
    <Timeline
      followNewItems={false}
      header={
        <>
          <div className="border-border/60 border-b-[0.5px] px-5 py-4">
            <Text fontWeight="medium" role="status">
              {results.isPending
                ? "Searching…"
                : `${matches.length}${results.hasNextPage ? "+" : ""} ${matches.length === 1 ? "result" : "results"}`}{" "}
              for “{query}”
            </Text>
            <Text className="mt-1" color="muted">
              Messages and replies in this conversation
            </Text>
          </div>
          {results.error ? (
            <div className="px-5 py-4">
              <Text>Could not search this conversation.</Text>
              <Button
                onClick={() => {
                  void results.refetch();
                }}
                size="sm"
              >
                Try again
              </Button>
            </div>
          ) : null}
          {!results.error && !results.isPending && matches.length === 0 ? (
            <div className="px-6 py-12 text-center">
              <Text fontWeight="medium">No matching messages</Text>
              <Text className="mt-2" color="muted">
                Try another word or phrase.
              </Text>
            </div>
          ) : null}
          {results.hasNextPage ? (
            <Button
              className="m-4"
              disabled={results.isFetchingNextPage}
              onClick={() => {
                void results.fetchNextPage();
              }}
              size="sm"
            >
              {results.isFetchingNextPage ? "Loading…" : "Load more results"}
            </Button>
          ) : null}
        </>
      }
      items={matches}
      key={`${conversation.id}:${query}`}
      label="Conversation search results"
      renderItem={(message) => (
        <div className="border-border/60 border-b-[0.5px] py-2">
          {message.threadRootId ? (
            <Text className="px-5 pt-2" color="muted">
              Reply in a thread
            </Text>
          ) : null}
          <MessageRow
            draftScope="search"
            mayaAvatarUrl={mayaAvatarUrl}
            message={message}
            onDeleteMessage={onDeleteMessage}
            onEditMessage={onEditMessage}
            onMarkUnread={onMarkUnread}
            onOpenResources={onOpenResources}
            onReply={(item, trigger) => {
              onReply(item, trigger);
            }}
            onToggleReaction={onToggleReaction}
          />
          <div className="px-5 pb-2">
            <Button
              color="tertiary"
              leftIcon={<ReplyIcon className="h-4 w-auto" />}
              onClick={() => {
                onViewMessage(message.id);
              }}
              size="sm"
              type="button"
              variant="naked"
            >
              {message.threadRootId ? "Open thread" : "View in conversation"}
            </Button>
          </div>
        </div>
      )}
    />
  );
};
