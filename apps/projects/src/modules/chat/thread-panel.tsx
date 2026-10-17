"use client";

import type { Ref } from "react";
import { CloseIcon } from "icons";
import { Button } from "ui";
import type { PreviewDraft } from "./chat-rich-text";
import { CHAT_SIDE_PANEL_CLASS, ICON_BUTTON_CLASS } from "./chat-styles";
import { Composer } from "./composer";
import { MessageRow } from "./message-row";
import type { ChatMessage } from "./types";
import { useChatIdentity } from "./chat-context";
import { TypingIndicator } from "./typing-indicator";
import { Timeline } from "./timeline";
import { AudioPlaybackBar } from "./audio-playback-bar";
import { ChatUnreadBadge } from "./unread-badge";

export const ThreadPanel = ({
  thread,
  replies,
  mayaAvatarUrl,
  closeButtonRef,
  allowMaya,
  draft,
  onClose,
  onDraftChange,
  onSend,
  onToggleReaction,
  onEditMessage,
  onDeleteMessage,
  onMarkUnread,
  onRetryMessage,
  hasOlder = false,
  hasNewer = false,
  loading = false,
  onLoadOlder,
  onLoadNewer,
  targetMessageId,
  scrollToMessageId,
  onVisibleMessage,
  onVisibleMessages,
  error,
  onRetry,
}: {
  thread: ChatMessage;
  replies: ChatMessage[];
  mayaAvatarUrl?: string | null;
  closeButtonRef: Ref<HTMLButtonElement>;
  allowMaya: boolean;
  draft: PreviewDraft;
  onClose: () => void;
  onDraftChange: (value: PreviewDraft) => void;
  onSend: () => void;
  onToggleReaction: (messageId: string, emoji: string) => void;
  onEditMessage: (
    messageId: string,
    draft: PreviewDraft,
  ) => void | Promise<void>;
  onDeleteMessage: (messageId: string) => void | Promise<void>;
  onMarkUnread?: (message: ChatMessage) => void;
  onRetryMessage?: (messageId: string) => void;
  hasOlder?: boolean;
  hasNewer?: boolean;
  loading?: boolean;
  onLoadOlder?: () => Promise<unknown>;
  onLoadNewer?: () => void;
  targetMessageId?: string | null;
  scrollToMessageId?: string | null;
  onVisibleMessage?: (message: ChatMessage) => void;
  onVisibleMessages?: (messages: readonly ChatMessage[]) => void;
  error?: Error | null;
  onRetry?: () => void;
}) => {
  const latestReply = replies.at(-1);
  const { userId } = useChatIdentity();
  return (
    <aside aria-label="Thread" className={CHAT_SIDE_PANEL_CLASS}>
      <header className="border-border flex h-16 shrink-0 items-center justify-between border-b-[0.5px] px-5">
        <div className="flex items-center gap-2.5">
          <h2 className="text-base font-semibold">Thread</h2>
          <ChatUnreadBadge
            count={thread.unreadReplyCount}
            label="unread replies"
          />
        </div>
        <div className="flex items-center gap-1">
          <Button
            aria-label="Close thread"
            asIcon
            className={ICON_BUTTON_CLASS}
            color="tertiary"
            onClick={onClose}
            ref={closeButtonRef}
            size="sm"
            variant="naked"
          >
            <CloseIcon className="size-4" />
          </Button>
        </div>
      </header>
      <div className="min-[1400px]:hidden">
        <AudioPlaybackBar />
      </div>
      <Timeline
        followOwnItemId={
          latestReply?.author.id === userId ? latestReply.id : undefined
        }
        hasOlder={hasOlder}
        header={
          <>
            <div className="pt-4">
              <MessageRow
                draftScope="thread"
                mayaAvatarUrl={mayaAvatarUrl}
                message={thread}
                onDeleteMessage={onDeleteMessage}
                onEditMessage={onEditMessage}
                onMarkUnread={onMarkUnread}
                onRetryMessage={onRetryMessage}
                onToggleReaction={onToggleReaction}
              />
            </div>
            <div className="text-text-muted mx-5 my-3 flex items-center gap-3 text-base">
              <span>
                {thread.replyCount}{" "}
                {thread.replyCount === 1 ? "reply" : "replies"}
              </span>
              <span className="bg-border/70 h-px flex-1" />
            </div>
            {hasOlder && loading ? (
              <p className="text-text-muted px-5 pb-2" role="status">
                Loading earlier replies…
              </p>
            ) : null}
            {error ? (
              <div className="px-5 py-3" role="alert">
                <p className="text-danger">{error.message}</p>
                <Button onClick={onRetry} size="sm">
                  Try again
                </Button>
              </div>
            ) : null}
            {replies.length === 0 && !loading && !error ? (
              <p className="text-text-muted px-5 py-3">
                Keep the conversation going with a reply.
              </p>
            ) : null}
          </>
        }
        isFetching={loading}
        items={replies}
        key={`timeline:${thread.id}`}
        label="Thread message history"
        onLoadOlder={onLoadOlder}
        onVisibleItem={onVisibleMessage}
        onVisibleItems={onVisibleMessages}
        paginationError={error}
        renderItem={(message) => (
          <MessageRow
            draftScope="thread"
            mayaAvatarUrl={mayaAvatarUrl}
            message={message}
            onDeleteMessage={onDeleteMessage}
            onEditMessage={onEditMessage}
            onMarkUnread={onMarkUnread}
            onRetryMessage={onRetryMessage}
            onToggleReaction={onToggleReaction}
          />
        )}
        scrollToItemId={scrollToMessageId}
        targetItemId={targetMessageId}
      />
      {hasNewer ? (
        <div className="px-5 py-2">
          <Button
            color="tertiary"
            disabled={loading}
            onClick={onLoadNewer}
            size="sm"
          >
            Load newer replies
          </Button>
        </div>
      ) : null}

      <Composer
        allowMaya={allowMaya}
        draft={draft}
        key={`composer:${thread.id}`}
        label="Reply in thread…"
        onDraftChange={onDraftChange}
        onSend={onSend}
        threadRootId={thread.id}
      />
      <TypingIndicator
        conversationId={thread.conversationId}
        messages={replies}
        threadRootId={thread.id}
      />
    </aside>
  );
};
