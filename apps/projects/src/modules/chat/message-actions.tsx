"use client";

import { useId, useRef, useState } from "react";
import {
  DeleteIcon,
  EditIcon,
  MessageReplyIcon,
  MoreHorizontalIcon,
} from "icons";
import { cn } from "lib";
import { Input, Menu, Popover, Tooltip } from "ui";
import {
  COMMENT_REACTION_GROUPS,
  COMMENT_REACTIONS,
} from "@/modules/story/types/comment-collaboration";
import type { MessageReplyHandler } from "./message-row";
import type { ChatMessage, ChatReaction } from "./types";
import { useChatIdentity } from "./chat-context";

const QUICK_REACTIONS = [
  { emoji: "✅", label: "Done" },
  { emoji: "👀", label: "Eyes" },
  { emoji: "🙌", label: "Raised hands" },
] as const;

const ACTION_CLASS =
  "text-text-muted hover:bg-foreground/[0.035] dark:hover:bg-foreground/[0.05] hover:text-foreground focus-visible:ring-primary flex size-9 shrink-0 items-center justify-center rounded-lg outline-none focus-visible:ring-2";

const AddReactionIcon = () => (
  <svg
    aria-hidden
    className="size-[18px]"
    fill="none"
    stroke="currentColor"
    strokeLinecap="round"
    strokeLinejoin="round"
    strokeWidth="1.6"
    viewBox="0 0 24 24"
  >
    <path d="M20 11a8 8 0 1 0-8 9M8 13s1.5 2 4 2c.9 0 1.6-.2 2.2-.5M18 15v6m-3-3h6" />
    <circle cx="8.5" cy="8.5" fill="currentColor" r=".5" />
    <circle cx="15.5" cy="8.5" fill="currentColor" r=".5" />
  </svg>
);

const ReactionPicker = ({
  author,
  open,
  onOpenChange,
  onSelect,
  reactions,
}: {
  author: string;
  open: boolean;
  onOpenChange: (open: boolean) => void;
  onSelect: (emoji: string) => void;
  reactions: ChatReaction[];
}) => {
  const [search, setSearch] = useState("");
  const selected = new Set(
    reactions.filter(({ reacted }) => reacted).map(({ emoji }) => emoji),
  );
  const choices = COMMENT_REACTIONS.filter(({ emoji, label }) =>
    `${label} ${emoji}`.toLowerCase().includes(search.trim().toLowerCase()),
  );
  const setOpen = (next: boolean) => {
    onOpenChange(next);
    if (!next) setSearch("");
  };

  return (
    <Popover onOpenChange={setOpen} open={open}>
      <Tooltip title="Add reaction">
        <Popover.Trigger asChild>
          <button
            aria-label={`Add reaction to ${author}'s message`}
            className={ACTION_CLASS}
            type="button"
          >
            <AddReactionIcon />
          </button>
        </Popover.Trigger>
      </Tooltip>
      <Popover.Content
        align="end"
        aria-label="Choose a reaction"
        className="bg-surface/95 dark:bg-background/95 w-80 max-w-[calc(100vw-2rem)] p-2"
      >
        <Input
          aria-label="Search reactions"
          onChange={(event) => {
            setSearch(event.target.value);
          }}
          placeholder="Search reactions…"
          size="sm"
          value={search}
        />
        <div className="mt-2 max-h-64 overflow-y-auto overscroll-contain">
          {COMMENT_REACTION_GROUPS.map((group) => {
            const groupChoices = choices.filter(
              (choice) => choice.group === group,
            );
            if (!groupChoices.length) return null;
            return (
              <div
                aria-label={group}
                className="mb-2 last:mb-0"
                key={group}
                role="group"
              >
                <p className="text-text-muted px-1 py-1">{group}</p>
                <div className="grid grid-cols-8 gap-0.5">
                  {groupChoices.map(({ emoji, label }) => (
                    <button
                      aria-label={label}
                      aria-pressed={selected.has(emoji)}
                      className={cn(
                        "hover:bg-foreground/[0.035] dark:hover:bg-foreground/[0.05] focus-visible:ring-primary text-foreground flex size-9 items-center justify-center rounded-lg text-xl opacity-100 outline-none focus-visible:ring-2",
                        selected.has(emoji) && "bg-primary/10",
                      )}
                      key={emoji}
                      onClick={() => {
                        onSelect(emoji);
                        setOpen(false);
                      }}
                      title={label}
                      type="button"
                    >
                      {emoji}
                    </button>
                  ))}
                </div>
              </div>
            );
          })}
        </div>
        {!choices.length ? (
          <p className="text-text-muted px-1 py-3">No matching reactions.</p>
        ) : null}
      </Popover.Content>
    </Popover>
  );
};

export const MessageActions = ({
  message,
  onReply,
  onToggleReaction,
  onEdit,
  onDelete,
}: {
  message: ChatMessage;
  onReply?: MessageReplyHandler;
  onToggleReaction: (messageId: string, emoji: string) => void;
  onEdit: () => void;
  onDelete: () => void;
}) => {
  const { userId } = useChatIdentity();
  const [expanded, setExpanded] = useState(false);
  const [pickerOpen, setPickerOpen] = useState(false);
  const [menuOpen, setMenuOpen] = useState(false);
  const actionSelected = useRef(false);
  const toolbarId = useId();
  const reactions = message.reactions;
  const selected = new Set(
    reactions.filter(({ reacted }) => reacted).map(({ emoji }) => emoji),
  );

  return (
    <div className="absolute -top-3 right-5 z-10 flex items-center gap-1">
      <div
        aria-label={`Actions for ${message.author.fullName}'s message`}
        className={cn(
          "bg-surface/95 dark:bg-background/95 border-border/70 pointer-events-none flex items-center gap-0.5 rounded-xl border-[0.5px] px-1.5 py-0.5 opacity-0 shadow-md backdrop-blur-md group-focus-within:pointer-events-auto group-focus-within:opacity-100 group-hover:pointer-events-auto group-hover:opacity-100",
          (expanded || pickerOpen || menuOpen) &&
            "pointer-events-auto opacity-100",
        )}
        id={toolbarId}
        role="group"
      >
        {QUICK_REACTIONS.map(({ emoji, label }) => (
          <Tooltip key={emoji} title={label}>
            <button
              aria-label={`${selected.has(emoji) ? "Remove" : "Add"} ${label.toLowerCase()} reaction`}
              aria-pressed={selected.has(emoji)}
              className={cn(
                ACTION_CLASS,
                "text-foreground text-[18px] opacity-100",
                selected.has(emoji) && "bg-primary/10",
              )}
              onClick={() => {
                onToggleReaction(message.id, emoji);
                setExpanded(false);
              }}
              type="button"
            >
              <span aria-hidden className="opacity-100">
                {emoji}
              </span>
            </button>
          </Tooltip>
        ))}
        <span aria-hidden className="bg-border mx-0.5 h-4 w-px" />
        <ReactionPicker
          author={message.author.fullName}
          onOpenChange={setPickerOpen}
          onSelect={(emoji) => {
            onToggleReaction(message.id, emoji);
            setExpanded(false);
          }}
          open={pickerOpen}
          reactions={reactions}
        />
        {onReply ? (
          <Tooltip title="Reply in thread">
            <button
              aria-label={`Reply to ${message.author.fullName}`}
              className={ACTION_CLASS}
              onClick={(event) => {
                onReply(message, event.currentTarget);
                setExpanded(false);
              }}
              type="button"
            >
              <MessageReplyIcon className="size-[18px]" />
            </button>
          </Tooltip>
        ) : null}
        {message.author.id === userId ? (
          <Menu onOpenChange={setMenuOpen} open={menuOpen}>
            <Tooltip title="More actions">
              <Menu.Button>
                <button
                  aria-label="More message actions"
                  className={ACTION_CLASS}
                  data-message-menu-trigger
                  type="button"
                >
                  <MoreHorizontalIcon className="size-[18px]" />
                </button>
              </Menu.Button>
            </Tooltip>
            <Menu.Items
              align="end"
              className="min-w-40"
              onCloseAutoFocus={(event) => {
                if (actionSelected.current) {
                  event.preventDefault();
                  actionSelected.current = false;
                }
              }}
            >
              <Menu.Group>
                <Menu.Item
                  onSelect={() => {
                    actionSelected.current = true;
                    onEdit();
                  }}
                >
                  <EditIcon className="size-4" /> Edit message
                </Menu.Item>
                <Menu.Item
                  className="text-danger"
                  onSelect={() => {
                    actionSelected.current = true;
                    onDelete();
                  }}
                >
                  <DeleteIcon className="size-4" /> Delete message
                </Menu.Item>
              </Menu.Group>
            </Menu.Items>
          </Menu>
        ) : null}
      </div>
      <button
        aria-controls={toolbarId}
        aria-expanded={expanded}
        aria-label={`Actions for ${message.author.fullName}'s message`}
        className={cn(
          ACTION_CLASS,
          "border-border bg-surface/95 dark:bg-background/95 size-7 rounded-lg border shadow-xs backdrop-blur-md [@media(hover:hover)]:hidden",
        )}
        onClick={() => {
          setExpanded((current) => !current);
        }}
        type="button"
      >
        <MoreHorizontalIcon className="size-4" />
      </button>
    </div>
  );
};
