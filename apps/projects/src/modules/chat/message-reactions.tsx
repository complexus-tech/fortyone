import { cn } from "lib";
import { Tooltip } from "ui";
import type { ChatReaction } from "./types";

export const MessageReactions = ({
  messageId,
  reactions,
  onToggleReaction,
}: {
  messageId: string;
  reactions: ChatReaction[];
  onToggleReaction: (messageId: string, emoji: string) => void;
}) => {
  const visibleReactions = reactions.filter(({ count }) => count > 0);
  if (!visibleReactions.length) return null;

  return (
    <div aria-label="Reactions" className="mt-1.5 flex flex-wrap gap-1">
      {visibleReactions.map(({ emoji, count, reacted = false, users }) => (
        <Tooltip
          key={emoji}
          title={
            users?.length ? (
              <div className="flex max-h-60 flex-col overflow-y-auto">
                {users.map((person) => (
                  <span key={person.id}>{person.fullName}</span>
                ))}
              </div>
            ) : undefined
          }
        >
          <button
            aria-label={`${reacted ? "Remove" : "Add"} ${emoji} reaction, ${count} ${count === 1 ? "reaction" : "reactions"}`}
            aria-pressed={reacted}
            className={cn(
              "border-border bg-surface/95 dark:bg-background/95 hover:bg-foreground/[0.025] dark:hover:bg-foreground/[0.035] focus-visible:ring-primary/40 inline-flex items-center gap-1.5 rounded-full border px-2 py-0.5 outline-none focus-visible:ring-2",
              reacted &&
                "border-primary/30 bg-primary/5 text-primary dark:bg-primary/10",
            )}
            onClick={() => {
              onToggleReaction(messageId, emoji);
            }}
            type="button"
          >
            <span aria-hidden className="text-foreground opacity-100">
              {emoji}
            </span>
            <span className="tabular-nums">{count}</span>
          </button>
        </Tooltip>
      ))}
    </div>
  );
};
