import { useId } from "react";
import { Avatar } from "ui";
import { cn } from "lib";
import type { Comment } from "@/types/comment";
import { FORMER_USER_NAME, isFormerUser } from "@/lib/former-user";
import type { CommentMayaPendingStatus } from "../types/comment-maya";
import { CommentMayaPendingLabel } from "./comment-maya-pending";

export const CommentReplySummary = ({
  replies,
  onToggle,
  threadId,
  isDialog,
  hasUnreadReplies,
  mayaStatus,
}: {
  replies: Comment[];
  onToggle: () => void;
  threadId: string;
  isDialog?: boolean;
  hasUnreadReplies: boolean;
  mayaStatus?: CommentMayaPendingStatus;
}) => {
  const unreadDescriptionId = useId();
  const participants = Array.from(
    new Map(replies.map((reply) => [reply.userId, reply])).values(),
  ).slice(0, 4);
  const countLabel = `${replies.length} ${replies.length === 1 ? "reply" : "replies"}`;
  return (
    <button
      aria-controls={`comment-replies-${threadId}`}
      aria-describedby={hasUnreadReplies ? unreadDescriptionId : undefined}
      aria-expanded={false}
      aria-label={
        mayaStatus
          ? "Show replies and Maya’s response"
          : `${countLabel}. Show replies`
      }
      className="text-text-muted hover:text-foreground focus-visible:ring-primary ml-9 inline-flex min-h-8 max-w-[calc(100%_-_2.25rem)] items-center gap-2 rounded-md py-1 pr-1 align-middle outline-none focus-visible:ring-2 [@media(pointer:coarse)]:min-h-11"
      onClick={onToggle}
      type="button"
    >
      {participants.length > 0 ? (
        <span className="flex shrink-0 items-center -space-x-1.5 px-[2px]">
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
              size="xs"
              src={isFormerUser(userId) ? null : user.avatarUrl}
            />
          ))}
        </span>
      ) : null}
      {mayaStatus ? (
        <CommentMayaPendingLabel
          className="min-w-0"
          compact
          status={mayaStatus}
        />
      ) : (
        <span className="text-[0.95rem] leading-5 whitespace-nowrap">
          {countLabel}
        </span>
      )}
      {hasUnreadReplies ? (
        <>
          <span
            aria-hidden
            className="bg-primary size-1.5 shrink-0 rounded-full"
          />
          <span className="sr-only" id={unreadDescriptionId}>
            New replies
          </span>
        </>
      ) : null}
    </button>
  );
};
