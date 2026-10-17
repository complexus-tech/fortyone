"use client";

import type { Ref } from "react";
import { useState } from "react";
import { ArrowLeftIcon, ChatIcon, LinkIcon } from "icons";
import { cn } from "lib";
import { Button, Tabs } from "ui";
import { PersonAvatar } from "./avatars";
import type { PreviewDraft } from "./chat-rich-text";
import { ICON_BUTTON_CLASS } from "./chat-styles";
import { AudioPlaybackBar } from "./audio-playback-bar";
import { Composer } from "./composer";
import {
  ConversationSearchHeader,
  ConversationSearchResults,
} from "./conversation-search";
import { FilesLinks } from "./files-links";
import type { MessageReplyHandler } from "./message-row";
import { MessageRow } from "./message-row";
import type { ConversationFeed, ChatMessage, ConversationTab } from "./types";
import { useChatIdentity } from "./chat-context";
import { getMessageDate } from "./format";
import { TypingIndicator } from "./typing-indicator";
import { Timeline } from "./timeline";

const getChatMessageDate = (message: ChatMessage) =>
  getMessageDate(message.createdAt);

export const ConversationView = ({
  conversation,
  headingRef,
  mayaAvatarUrl,
  showConversation,
  hasThread,
  draft,
  onBack,
  onReply,
  onDraftChange,
  onSend,
  activeTab,
  onTabChange,
  onViewMessage,
  targetMessageId,
  scrollToMessageId,
  onToggleReaction,
  onEditMessage,
  onDeleteMessage,
  sending = false,
  hasOlder = false,
  hasNewer = false,
  loadingOlder = false,
  fetchingHistory = false,
  historyError,
  onLoadOlder,
  onJumpToLatest,
  onVisibleMessage,
}: {
  conversation: ConversationFeed;
  headingRef: Ref<HTMLHeadingElement>;
  mayaAvatarUrl?: string | null;
  showConversation: boolean;
  hasThread: boolean;
  draft: PreviewDraft;
  onBack: () => void;
  onReply: MessageReplyHandler;
  onDraftChange: (value: PreviewDraft) => void;
  onSend: () => void;
  activeTab: ConversationTab;
  onTabChange: (value: ConversationTab) => void;
  onViewMessage: (messageId: string) => void;
  targetMessageId: string | null;
  scrollToMessageId?: string | null;
  onToggleReaction: (messageId: string, emoji: string) => void;
  onEditMessage: (
    messageId: string,
    draft: PreviewDraft,
  ) => void | Promise<void>;
  onDeleteMessage: (messageId: string) => void | Promise<void>;
  sending?: boolean;
  hasOlder?: boolean;
  hasNewer?: boolean;
  loadingOlder?: boolean;
  fetchingHistory?: boolean;
  historyError?: Error | null;
  onLoadOlder?: () => Promise<unknown>;
  onJumpToLatest?: () => void;
  onVisibleMessage?: (message: ChatMessage) => void;
}) => {
  const [search, setSearch] = useState("");
  const { userId, mayaEnabled } = useChatIdentity();
  const latestMessage = conversation.messages.at(-1);

  return (
    <section
      aria-label={`${conversation.name} conversation`}
      className={cn(
        "@container/conversation flex min-w-0 flex-1 flex-col",
        !showConversation && "hidden lg:flex",
        hasThread && "hidden min-[1400px]:flex",
      )}
    >
      <h2 className="sr-only" ref={headingRef} tabIndex={-1}>
        {conversation.name}
      </h2>
      <Tabs
        className="flex min-h-0 flex-1 flex-col"
        onValueChange={(value) => {
          if (value === "messages" || value === "files") onTabChange(value);
        }}
        value={activeTab}
      >
        <ConversationSearchHeader
          leading={
            <div className="flex min-w-0 items-center justify-between gap-2">
              <div className="flex min-w-0 items-center gap-2">
                <Button
                  aria-label="Back to conversations"
                  asIcon
                  className={cn(ICON_BUTTON_CLASS, "lg:hidden")}
                  color="tertiary"
                  onClick={onBack}
                  size="sm"
                  variant="naked"
                >
                  <ArrowLeftIcon className="size-4" />
                </Button>
                <Tabs.List
                  aria-label="Conversation content"
                  className="mx-0 flex-nowrap md:mx-0"
                >
                  <Tabs.Tab
                    leftIcon={<ChatIcon className="h-[1.1rem] w-auto" />}
                    value="messages"
                  >
                    Messages
                  </Tabs.Tab>
                  <Tabs.Tab
                    leftIcon={<LinkIcon className="h-[1.1rem] w-auto" />}
                    value="files"
                  >
                    Files & links
                  </Tabs.Tab>
                </Tabs.List>
              </div>
              <div className="flex shrink-0 items-center gap-3">
                {conversation.kind === "team" ? (
                  <div
                    aria-label="Team members"
                    className="isolate hidden items-center -space-x-1.5 @min-[460px]/conversation:flex"
                  >
                    {conversation.members.slice(0, 4).map((person) => (
                      <span
                        className="bg-surface ring-surface relative inline-flex rounded-full ring-2"
                        key={person.id}
                      >
                        <PersonAvatar
                          mayaAvatarUrl={mayaAvatarUrl}
                          person={person}
                        />
                      </span>
                    ))}
                    <span className="text-text-muted ml-3 pl-2">
                      {conversation.memberCount}
                    </span>
                  </div>
                ) : null}
              </div>
            </div>
          }
          onActivate={() => {
            onTabChange("messages");
          }}
          onSearch={(value) => {
            setSearch(value);
            onTabChange("messages");
          }}
          query={search}
        />
        <AudioPlaybackBar />
        <Tabs.Panel
          aria-label={search ? "Conversation search results" : "Messages"}
          className="flex min-h-0 flex-1 flex-col outline-none"
          value="messages"
        >
          {search ? (
            <ConversationSearchResults
              conversation={conversation}
              mayaAvatarUrl={mayaAvatarUrl}
              onDeleteMessage={onDeleteMessage}
              onEditMessage={onEditMessage}
              onOpenResources={() => {
                setSearch("");
                onTabChange("files");
              }}
              onReply={onReply}
              onToggleReaction={onToggleReaction}
              onViewMessage={(messageId) => {
                setSearch("");
                onViewMessage(messageId);
              }}
              query={search}
            />
          ) : (
            <>
              <Timeline
                followOwnItemId={
                  latestMessage?.author.id === userId
                    ? latestMessage.id
                    : undefined
                }
                getItemDate={getChatMessageDate}
                hasOlder={hasOlder}
                header={
                  hasOlder ? (
                    <div className="text-text-muted flex min-h-10 items-center justify-center gap-2 px-5 py-2">
                      {loadingOlder ? (
                        <span role="status">Loading earlier messages…</span>
                      ) : null}
                      {historyError ? (
                        <div role="alert">
                          <span>Could not load earlier messages.</span>
                          <Button
                            color="tertiary"
                            onClick={() => {
                              void onLoadOlder?.();
                            }}
                            size="sm"
                          >
                            Try again
                          </Button>
                        </div>
                      ) : null}
                    </div>
                  ) : (
                    <div className="px-5 pt-7 pb-5">
                      <h3 className="text-xl font-semibold tracking-tight">
                        {conversation.kind === "team"
                          ? `# ${conversation.name}`
                          : `You and ${conversation.name.split(" ")[0]}`}
                      </h3>
                      <p className="text-text-muted mt-1 text-base leading-6">
                        {conversation.description}
                      </p>
                    </div>
                  )
                }
                isActive={!hasThread}
                isFetching={fetchingHistory}
                items={conversation.messages}
                key={`timeline:${conversation.id}`}
                onLoadOlder={onLoadOlder}
                onVisibleItem={onVisibleMessage}
                paginationError={historyError}
                renderItem={(message) => (
                  <div
                    className={cn(
                      "outline-none",
                      targetMessageId === message.id && "bg-primary/5",
                    )}
                    id={`chat-message-${message.id}`}
                    tabIndex={-1}
                  >
                    <MessageRow
                      mayaAvatarUrl={mayaAvatarUrl}
                      message={message}
                      onDeleteMessage={onDeleteMessage}
                      onEditMessage={onEditMessage}
                      onOpenResources={() => {
                        onTabChange("files");
                      }}
                      onReply={onReply}
                      onToggleReaction={onToggleReaction}
                    />
                  </div>
                )}
                scrollToItemId={scrollToMessageId}
                startAtEnd
                targetItemId={targetMessageId}
              />
              {hasNewer ? (
                <div className="flex justify-center py-2">
                  <Button color="tertiary" onClick={onJumpToLatest} size="sm">
                    Jump to latest messages
                  </Button>
                </div>
              ) : null}

              <Composer
                allowMaya={mayaEnabled}
                disabled={sending}
                draft={draft}
                key={`composer:${conversation.id}`}
                label={`Message ${conversation.name}…`}
                onDraftChange={onDraftChange}
                onSend={onSend}
              />
              <TypingIndicator
                active={!hasThread}
                conversationId={conversation.id}
                messages={conversation.messages}
              />
            </>
          )}
        </Tabs.Panel>
        <Tabs.Panel
          aria-label="Files & links"
          className="flex min-h-0 flex-1 flex-col outline-none"
          value="files"
        >
          <FilesLinks
            conversation={conversation}
            key={conversation.id}
            onViewMessage={onViewMessage}
          />
        </Tabs.Panel>
      </Tabs>
    </section>
  );
};
