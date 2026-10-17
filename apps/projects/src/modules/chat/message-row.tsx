"use client";

import { useRef, useState } from "react";
import { cn } from "lib";
import { PersonAvatar } from "./avatars";
import { ChatPlainText, ChatRichText } from "./chat-rich-text";
import type { PreviewDraft } from "./chat-rich-text";
import { Composer } from "./composer";
import { MessageActions } from "./message-actions";
import { MessageDeleteDialog } from "./message-delete-dialog";
import type { MessageDraftScope } from "./message-edit-drafts";
import { useMessageEditDraft } from "./message-edit-drafts";
import type { ChatMessage } from "./types";
import { useChatIdentity } from "./chat-context";
import { ChatTaskResults } from "./chat-task-results";
import { ChatMayaStatus } from "./chat-maya-status";
import { formatMessageTime } from "./format";
import { MessageAttachments } from "./message-attachments";
import { MessageLinkPreviews } from "./link-preview";
import { ChatUnreadBadge } from "./unread-badge";

const restoreMessageActionsFocus = (row: HTMLElement | null) => {
  const surface = row?.closest<HTMLElement>(
    'section[aria-label$=" conversation"], aside[aria-label="Thread"]',
  );
  requestAnimationFrame(() => {
    if (row?.isConnected) {
      const trigger = row.querySelector<HTMLButtonElement>(
        "[data-message-menu-trigger]",
      );
      (trigger ?? row).focus();
      return;
    }
    if (!surface?.isConnected) return;
    const search = Array.from(
      surface.querySelectorAll<HTMLInputElement>("input"),
    ).find((input) => !input.disabled && input.getClientRects().length > 0);
    const fallback =
      search ??
      surface.querySelector<HTMLButtonElement>(
        'button[aria-label="Close thread"]',
      ) ??
      surface.querySelector<HTMLElement>('h2[tabindex="-1"]');
    fallback?.focus();
  });
};

export type MessageReplyHandler = (
  message: ChatMessage,
  trigger: HTMLButtonElement,
) => void;

const MessageBody = ({
  message,
  editDraft,
  onDraftChange,
  onSave,
  onCancel,
  autoFocus,
  pending,
}: {
  message: ChatMessage;
  editDraft: PreviewDraft | null;
  onDraftChange: (draft: PreviewDraft) => void;
  onSave: () => void;
  onCancel: () => void;
  autoFocus: boolean;
  pending: boolean;
}) => {
  if (message.deletedAt) {
    return (
      <p className="text-text-muted text-base leading-6 italic">
        This message was deleted.
      </p>
    );
  }
  if (editDraft) {
    return (
      <Composer
        allowMaya
        autoFocus={autoFocus}
        disabled={pending}
        draft={editDraft}
        intent="edit"
        label="Edit message"
        onCancel={onCancel}
        onDraftChange={onDraftChange}
        onSend={onSave}
      />
    );
  }
  if (message.richContent) {
    return <ChatRichText content={message.richContent} />;
  }
  return <ChatPlainText text={message.text} />;
};

const MessageDetails = ({
  message,
  onToggleReaction,
}: {
  message: ChatMessage;
  onToggleReaction: (messageId: string, emoji: string) => void;
}) => {
  if (message.deletedAt) return null;
  const reactions = message.reactions.filter(({ count }) => count > 0);
  return (
    <>
      {message.attachments.length > 0 ? (
        <MessageAttachments message={message} />
      ) : null}
      {message.links.length > 0 ? (
        <MessageLinkPreviews
          hasAttachments={message.attachments.length > 0}
          links={message.links}
        />
      ) : null}
      {message.taskRefs.length > 0 ? (
        <ChatTaskResults tasks={message.taskRefs} />
      ) : null}
      {reactions.length > 0 ? (
        <div aria-label="Reactions" className="mt-1.5 flex flex-wrap gap-1">
          {reactions.map(({ emoji, count, reacted = false }) => (
            <button
              aria-label={`${reacted ? "Remove" : "Add"} ${emoji} reaction, ${count} ${count === 1 ? "reaction" : "reactions"}`}
              aria-pressed={reacted}
              className={cn(
                "border-border bg-surface/95 dark:bg-background/95 hover:bg-foreground/[0.025] dark:hover:bg-foreground/[0.035] focus-visible:ring-primary/40 inline-flex items-center gap-1.5 rounded-full border px-2 py-0.5 outline-none focus-visible:ring-2",
                reacted &&
                  "border-primary/30 bg-primary/5 text-primary dark:bg-primary/10",
              )}
              key={emoji}
              onClick={() => {
                onToggleReaction(message.id, emoji);
              }}
              type="button"
            >
              <span aria-hidden className="text-foreground opacity-100">
                {emoji}
              </span>
              <span className="tabular-nums">{count}</span>
            </button>
          ))}
        </div>
      ) : null}
    </>
  );
};

export const MessageRow = ({
  message,
  mayaAvatarUrl,
  onReply,
  onToggleReaction,
  onEditMessage,
  onDeleteMessage,
  draftScope = "conversation",
}: {
  message: ChatMessage;
  mayaAvatarUrl?: string | null;
  onReply?: MessageReplyHandler;
  onOpenResources?: () => void;
  onToggleReaction: (messageId: string, emoji: string) => void;
  onEditMessage: (
    messageId: string,
    draft: PreviewDraft,
  ) => void | Promise<void>;
  onDeleteMessage: (messageId: string) => void | Promise<void>;
  draftScope?: MessageDraftScope;
}) => {
  const {
    draft: editDraft,
    setDraft: setEditDraft,
    clearMessageDrafts,
  } = useMessageEditDraft(message.id, draftScope);
  const [focusEditor, setFocusEditor] = useState(false);
  const [deleteOpen, setDeleteOpen] = useState(false);
  const [isSaving, setIsSaving] = useState(false);
  const [isDeleting, setIsDeleting] = useState(false);
  const rowRef = useRef<HTMLElement>(null);
  const { userId } = useChatIdentity();
  const isOwnMessage = message.author.id === userId;
  const restoreActionsFocus = () => {
    restoreMessageActionsFocus(rowRef.current);
  };
  const closeEditor = () => {
    setEditDraft(null);
    setFocusEditor(false);
    restoreActionsFocus();
  };

  return (
    <article
      className="group hover:bg-foreground/[0.025] focus-within:bg-foreground/[0.025] dark:hover:bg-foreground/[0.035] dark:focus-within:bg-foreground/[0.035] relative px-5 py-2 outline-none"
      ref={rowRef}
      tabIndex={-1}
    >
      <div className="flex w-full max-w-[max(56rem,92%)] min-w-0 gap-2.5">
        <PersonAvatar mayaAvatarUrl={mayaAvatarUrl} person={message.author} />
        <div className="min-w-0 flex-1">
          <div className="flex items-baseline gap-2">
            <span className="text-base leading-6 font-bold">
              {message.author.fullName}
            </span>
            <span className="text-text-muted">
              {formatMessageTime(message.createdAt)}
            </span>
            {message.editedAt && !message.deletedAt ? (
              <span className="text-text-muted">(edited)</span>
            ) : null}
          </div>
          <MessageBody
            autoFocus={focusEditor}
            editDraft={editDraft}
            message={message}
            onCancel={closeEditor}
            onDraftChange={setEditDraft}
            onSave={() => {
              if (!editDraft) return;
              setIsSaving(true);
              void Promise.resolve(onEditMessage(message.id, editDraft))
                .then(() => {
                  clearMessageDrafts();
                  closeEditor();
                })
                .catch(() => {
                  /* The mutation exposes the error and keeps this draft. */
                })
                .finally(() => {
                  setIsSaving(false);
                });
            }}
            pending={isSaving}
          />
          <MessageDetails
            message={message}
            onToggleReaction={onToggleReaction}
          />
          <ChatMayaStatus
            conversationId={message.conversationId}
            sourceMessageId={message.id}
          />
          {message.replyCount > 0 && onReply ? (
            <button
              aria-label={`Open thread with ${message.replyCount} replies${message.unreadReplyCount ? `, ${message.unreadReplyCount} unread` : ""}`}
              className="focus-visible:ring-primary/40 mt-1.5 flex max-w-full items-center gap-2 rounded px-0.5 py-1 text-left outline-none focus-visible:ring-2"
              onClick={(event) => {
                onReply(message, event.currentTarget);
              }}
              type="button"
            >
              <span className="[&>span]:ring-surface flex -space-x-1.5 [&>span]:rounded-full [&>span]:ring-2">
                {message.replyAuthors.slice(0, 3).map((person) => (
                  <PersonAvatar
                    key={person.id}
                    mayaAvatarUrl={mayaAvatarUrl}
                    person={person}
                    small
                  />
                ))}
              </span>
              <span className="text-primary shrink-0 text-base leading-5 font-medium">
                {message.replyCount}{" "}
                {message.replyCount === 1 ? "reply" : "replies"}
              </span>
              <ChatUnreadBadge
                count={message.unreadReplyCount}
                label="unread replies"
              />
              {message.lastReplyAt ? (
                <span className="text-text-muted group-hover:text-foreground group-focus-within:text-foreground truncate leading-5">
                  Last reply {formatMessageTime(message.lastReplyAt)}
                </span>
              ) : null}
            </button>
          ) : null}
        </div>
      </div>
      {!message.deletedAt && !editDraft ? (
        <MessageActions
          message={message}
          onDelete={() => {
            if (isOwnMessage) setDeleteOpen(true);
          }}
          onEdit={() => {
            if (isOwnMessage) {
              setFocusEditor(true);
              setEditDraft({
                text: message.text,
                content: message.richContent ?? null,
                attachments: message.attachments,
                baseVersion: message.version,
              });
            }
          }}
          onReply={onReply}
          onToggleReaction={onToggleReaction}
        />
      ) : null}
      {isOwnMessage ? (
        <MessageDeleteDialog
          hasReplies={message.replyCount > 0}
          onClose={() => {
            setDeleteOpen(false);
            restoreActionsFocus();
          }}
          onDelete={() => {
            setIsDeleting(true);
            void Promise.resolve(onDeleteMessage(message.id))
              .then(() => {
                setDeleteOpen(false);
                clearMessageDrafts();
              })
              .catch(() => {
                /* The mutation displays the error; keep confirmation open. */
              })
              .finally(() => {
                setIsDeleting(false);
              });
          }}
          open={deleteOpen}
          pending={isDeleting}
          text={message.text}
        />
      ) : null}
    </article>
  );
};
