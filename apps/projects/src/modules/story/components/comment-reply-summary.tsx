import { useId } from "react";
import { Avatar } from "ui";
import { cn } from "lib";
import type { Comment } from "@/types/comment";
import { FORMER_USER_NAME, isFormerUser } from "@/lib/former-user";

export const CommentReplySummary = ({
  replies,
  expanded,
  onToggle,
  threadId,
  isDialog,
  hasUnreadReplies,
}: {
  replies: Comment[];
  expanded: boolean;
  onToggle: () => void;
  threadId: string;
  isDialog?: boolean;
  hasUnreadReplies: boolean;
}) => {
  const unreadDescriptionId = useId();
  const participants = Array.from(
    new Map(replies.map((reply) => [reply.userId, reply])).values(),
  ).slice(0, 3);
  const countLabel = `${replies.length} ${replies.length === 1 ? "reply" : "replies"}`;
  return (
    <button
      aria-controls={`comment-replies-${threadId}`}
      aria-describedby={hasUnreadReplies ? unreadDescriptionId : undefined}
      aria-expanded={expanded}
      aria-label={`${countLabel}. ${expanded ? "Hide replies" : "Show replies"}`}
      className="text-text-muted hover:text-foreground focus-visible:ring-primary ml-9 inline-flex min-h-8 items-center gap-2 rounded-md outline-none focus-visible:ring-2 [@media(pointer:coarse)]:min-h-11"
      onClick={onToggle}
      type="button"
    >
      <span className="flex -space-x-2">
        {participants.map(({ userId, user }) => (
          <Avatar
            className={cn(
              "ring-surface ring-2",
              isDialog && "dark:ring-surface-elevated",
            )}
            key={userId}
            name={
              isFormerUser(userId)
                ? FORMER_USER_NAME
                : user.fullName || user.username
            }
            size="sm"
            src={isFormerUser(userId) ? null : user.avatarUrl}
          />
        ))}
      </span>
      <span className="text-sm">{countLabel}</span>
      {hasUnreadReplies ? (
        <>
          <span aria-hidden className="bg-primary size-1.5 rounded-full" />
          <span className="sr-only" id={unreadDescriptionId}>
            New replies
          </span>
        </>
      ) : null}
    </button>
  );
};
