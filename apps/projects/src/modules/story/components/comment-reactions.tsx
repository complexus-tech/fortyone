import { useState } from "react";
import { Input, Popover } from "ui";
import { cn } from "lib";
import {
  COMMENT_REACTION_GROUPS,
  COMMENT_REACTIONS,
} from "../types/comment-collaboration";
import type {
  CommentEmoji,
  CommentReaction,
} from "../types/comment-collaboration";

const AddReactionIcon = () => (
  <svg
    aria-hidden
    className="size-[16px]"
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

export const CommentReactions = ({
  reactions,
  disabled,
  onReact,
}: {
  reactions: CommentReaction[];
  disabled: boolean;
  onReact: (emoji: CommentEmoji, reacted: boolean) => void;
}) => {
  const [open, setOpen] = useState(false);
  const [search, setSearch] = useState("");
  const counts = new Map(
    reactions.map((reaction) => [reaction.emoji, reaction]),
  );
  const visible = COMMENT_REACTIONS.filter(
    ({ emoji }) => (counts.get(emoji)?.count ?? 0) > 0,
  );
  const choices = COMMENT_REACTIONS.filter(({ emoji, label }) =>
    `${label} ${emoji}`.toLowerCase().includes(search.trim().toLowerCase()),
  );
  const toggle = (emoji: CommentEmoji) => {
    onReact(emoji, !counts.get(emoji)?.reacted);
  };
  return (
    <div
      aria-label="Reactions"
      className="inline-flex flex-wrap items-center gap-1"
    >
      {visible.map(({ emoji, label }) => {
        const reaction = counts.get(emoji);
        const count = reaction?.count ?? 0;
        return (
          <button
            aria-label={`${label}, ${count} ${count === 1 ? "reaction" : "reactions"}`}
            aria-pressed={reaction?.reacted ?? false}
            className={cn(
              "text-text-muted hover:bg-state-hover focus-visible:ring-primary border-border inline-flex h-7 min-w-7 items-center justify-center gap-1.5 rounded-full border px-2 text-[0.95rem] outline-none focus-visible:ring-2 disabled:opacity-40",
              reaction?.reacted &&
                "border-primary/30 bg-primary/5 text-foreground",
            )}
            disabled={disabled}
            key={emoji}
            onClick={() => {
              toggle(emoji);
            }}
            title={label}
            type="button"
          >
            <span aria-hidden className="leading-none">
              {emoji}
            </span>
            <span className="tabular-nums">{count}</span>
          </button>
        );
      })}
      <Popover
        onOpenChange={(next) => {
          setOpen(next);
          if (!next) setSearch("");
        }}
        open={open}
      >
        <Popover.Trigger asChild>
          <button
            aria-label="Add reaction"
            className="comment-add-reaction text-text-muted border-border hover:bg-state-hover focus-visible:ring-primary flex size-[24px] items-center justify-center rounded-full border leading-none outline-none focus-visible:ring-2 disabled:opacity-40"
            disabled={disabled}
            type="button"
          >
            <AddReactionIcon />
          </button>
        </Popover.Trigger>
        <Popover.Content
          align="start"
          aria-label="Choose a reaction"
          className="w-80 max-w-[calc(100vw-2rem)] p-2"
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
                  <p className="text-text-muted px-1 py-1 text-sm">{group}</p>
                  <div className="grid grid-cols-8 gap-0.5">
                    {groupChoices.map(({ emoji, label }) => (
                      <button
                        aria-label={label}
                        aria-pressed={counts.get(emoji)?.reacted ?? false}
                        className={cn(
                          "hover:bg-state-hover focus-visible:ring-primary flex h-9 w-full items-center justify-center rounded-md text-xl outline-none focus-visible:ring-2",
                          counts.get(emoji)?.reacted && "bg-state-selected",
                        )}
                        disabled={disabled}
                        key={emoji}
                        onClick={() => {
                          toggle(emoji);
                          setOpen(false);
                          setSearch("");
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
            <p className="text-text-muted px-1 py-3 text-sm">
              No matching reactions.
            </p>
          ) : null}
        </Popover.Content>
      </Popover>
    </div>
  );
};
