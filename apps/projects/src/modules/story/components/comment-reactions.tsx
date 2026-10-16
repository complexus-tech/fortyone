import { cn } from "lib";
import { COMMENT_REACTIONS } from "../types/comment-collaboration";
import type {
  CommentEmoji,
  CommentReaction,
} from "../types/comment-collaboration";

export const CommentReactions = ({
  reactions,
  disabled,
  onReact,
}: {
  reactions: CommentReaction[];
  disabled: boolean;
  onReact: (emoji: CommentEmoji, reacted: boolean) => void;
}) => {
  const visible = COMMENT_REACTIONS.flatMap(({ emoji, label }) => {
    const reaction = reactions.find(
      (item) => item.emoji === emoji && item.count > 0,
    );
    return reaction ? [{ ...reaction, label }] : [];
  });
  if (!visible.length) return null;
  return (
    <div aria-label="Reactions" className="mt-2 ml-10 flex flex-wrap gap-1.5">
      {visible.map(({ emoji, label, count, reacted }) => (
        <button
          aria-label={`${label}, ${count} ${count === 1 ? "reaction" : "reactions"}`}
          aria-pressed={reacted}
          className={cn(
            "border-border hover:bg-surface-muted focus-visible:ring-primary inline-flex min-h-7 items-center gap-1.5 rounded-md border px-2 text-xs outline-none focus-visible:ring-2 disabled:opacity-50",
            reacted && "border-primary/30 bg-primary/5 text-primary",
          )}
          disabled={disabled}
          key={emoji}
          onClick={() => {
            onReact(emoji, !reacted);
          }}
          type="button"
        >
          <span aria-hidden>{emoji}</span>
          <span>{count}</span>
        </button>
      ))}
    </div>
  );
};
