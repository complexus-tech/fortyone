import { useState } from "react";
import { Box, Text, TimeAgo } from "ui";
import { CheckIcon } from "icons";
import { cn } from "lib";
import type { Comment } from "@/types/comment";
import { useDeleteCommentMutation } from "@/lib/hooks/delete-comment-mutation";
import { useCommentCollaboration } from "../hooks/comment-collaboration";
import { useCommentThreadReadTracking } from "../hooks/comment-read-tracking";
import { useCommentSeen } from "../hooks/use-comment-seen";
import {
  flattenComments,
  isUnreadComment,
  latestComment,
} from "../lib/comment-navigation";
import { MAX_COMMENT_DEPTH } from "../lib/comment-tree";
import type { CommentState } from "../types/comment-collaboration";
import type { CommentMayaRun } from "../types/comment-maya";
import type { MayaPendingReply } from "../hooks/use-maya-pending-replies";
import { CommentInput } from "./comment-input";
import { CommentAuthor } from "./comment-author";
import { CommentActions } from "./comment-actions";
import { CommentFooter } from "./comment-footer";
import { CommentBody } from "./comment-body";
import { CommentReplySummary } from "./comment-reply-summary";
import { CommentDeleteDialog } from "./comment-delete-dialog";
import { CommentMayaApproval } from "./comment-maya-approval";
import { CommentMayaPending } from "./comment-maya-pending";

type CommentRowProps = {
  comment: Comment;
  storyId: string;
  teamId: string;
  isDialog?: boolean;
  depth?: number;
  threadResolved?: boolean;
  currentUserId?: string;
  stateById: ReadonlyMap<string, CommentState>;
  targetId: string | null;
  highlightedId?: string | null;
  revealKey?: string | null;
  lastReadAt: string | null | undefined;
  readObservedAt?: string;
  mayaRuns?: CommentMayaRun[];
  mayaPending?: MayaPendingReply[];
  onCheckMaya?: () => void;
  onRetryMaya?: (run: CommentMayaRun) => void;
  checkingMaya?: boolean;
  retryingMayaRunId?: string;
};

export const CommentRow = ({
  comment: {
    id,
    userId,
    user,
    createdAt,
    updatedAt,
    comment,
    subComments,
    desktopPending,
  },
  storyId,
  teamId,
  isDialog,
  depth = 1,
  threadResolved = false,
  currentUserId,
  stateById,
  targetId,
  highlightedId = null,
  revealKey = targetId,
  lastReadAt,
  readObservedAt,
  mayaRuns = [],
  mayaPending = [],
  onCheckMaya,
  onRetryMaya,
  checkingMaya,
  retryingMayaRunId,
}: CommentRowProps) => {
  const { mutate: deleteComment, isPending: isDeleting } =
    useDeleteCommentMutation();
  const { reaction, resolution } = useCommentCollaboration(storyId, id);
  const [isOpen, setIsOpen] = useState(false);
  const [isReplying, setIsReplying] = useState(false);
  const [isEditing, setIsEditing] = useState(false);
  const isSubComment = depth > 1;
  const replies = flattenComments(subComments);
  const hasReplies = replies.length > 0 || mayaPending.length > 0;
  const descendantIds = new Set(replies.map((reply) => reply.id));
  const pendingHere = isSubComment
    ? mayaPending
    : mayaPending.filter(({ run }) => !descendantIds.has(run.sourceId));
  const [expansion, setExpansion] = useState<{
    revealKey: string | null;
    expanded: boolean;
  }>();
  const hasTargetReply = replies.some((reply) => reply.id === targetId);
  // One disclosure controls the whole thread, including nested replies. This
  // keeps the root's read receipt from consuming independently hidden branches.
  const rootExpanded =
    hasTargetReply && expansion?.revealKey !== revealKey
      ? true
      : expansion?.expanded ?? false;
  const showReplies = isSubComment ? hasReplies : rootExpanded;
  const threadRef = useCommentThreadReadTracking(
    storyId,
    id,
    !isSubComment && showReplies,
    latestComment(subComments)?.createdAt,
    readObservedAt,
    lastReadAt,
  );
  const isPending =
    Boolean(desktopPending) || id.startsWith("pending-comment:");
  const seenRef = useCommentSeen(
    storyId,
    id,
    user.isSystem && !isPending && !isEditing,
  );
  const state = stateById.get(id);
  const resolved = Boolean(state?.resolvedAt);
  const reactions = state?.reactions ?? [];
  const disabled =
    isPending || isDeleting || reaction.isPending || resolution.isPending;
  const handleCancel = () => {
    setIsReplying(false);
    setIsEditing(false);
  };
  const approval = mayaRuns.find(
    (run) => run.replyCommentId === id && run.status === "awaiting_approval",
  );
  const canReply =
    depth < MAX_COMMENT_DEPTH &&
    !isReplying &&
    !isEditing &&
    !resolved &&
    !threadResolved;
  const canCollapse = !isSubComment && hasReplies && showReplies;
  const handleReact = (
    emoji: Parameters<typeof reaction.mutate>[0]["emoji"],
    reacted: boolean,
  ) => {
    reaction.mutate({ emoji, reacted });
  };

  return (
    <Box
      className={cn(
        "comment-row relative scroll-mt-24 rounded-md pb-3 transition-[background-color,box-shadow] duration-300 outline-none motion-reduce:transition-none",
        {
          "ml-9 pt-1 pb-0": isSubComment,
          "bg-primary/5 ring-primary/20 ring-1": highlightedId === id,
        },
      )}
      data-comment-depth={depth}
      id={`comment-${id}`}
      tabIndex={-1}
    >
      <Box className="relative">
        {hasReplies ? (
          <span
            aria-hidden
            className={cn(
              "border-border/70 pointer-events-none absolute top-8 bottom-0 left-3.5 border-l",
              { "bottom-4 w-[1.35rem] rounded-bl-lg border-b": !showReplies },
            )}
          />
        ) : null}
        <Box
          className="group flex flex-wrap items-center gap-1"
          data-comment-anchor
        >
          <CommentAuthor isDialog={isDialog} user={user} userId={userId} />
          <Text className="mx-0.5 text-[0.95rem]" color="muted">
            ·
          </Text>
          <Text className="text-[0.95rem]" color="muted">
            <TimeAgo timestamp={createdAt} />
          </Text>
          {Date.parse(updatedAt) > Date.parse(createdAt) ? (
            <Text
              className="ml-1 text-[0.95rem]"
              color="muted"
              title={`Edited ${new Date(updatedAt).toUTCString()}`}
            >
              (edited)
            </Text>
          ) : null}
          {resolved ? (
            <Text
              className="ml-2 flex items-center gap-1 text-[0.95rem]"
              color="muted"
            >
              <CheckIcon className="size-3.5" /> Resolved
            </Text>
          ) : null}
          <CommentActions
            commentId={id}
            disabled={disabled}
            isOwner={userId === currentUserId}
            isRoot={!isSubComment}
            onDelete={() => {
              setIsOpen(true);
            }}
            onEdit={() => {
              setIsEditing(true);
            }}
            onResolve={() => {
              resolution.mutate(!resolved);
            }}
            resolved={resolved}
            storyId={storyId}
          />
        </Box>
        {!isEditing ? (
          <CommentBody
            className="mt-0.5 ml-9"
            html={comment}
            visibilityRef={user.isSystem ? seenRef : undefined}
          />
        ) : null}
        {approval ? (
          <CommentMayaApproval
            currentUserId={currentUserId}
            run={approval}
            storyId={storyId}
          />
        ) : null}
        {hasReplies && !showReplies ? (
          <Box className="mt-2">
            <CommentReplySummary
              hasUnreadReplies={replies.some((reply) =>
                isUnreadComment(reply, lastReadAt, currentUserId),
              )}
              isDialog={isDialog}
              mayaStatus={mayaPending[0]?.status}
              onToggle={() => {
                setExpansion({ revealKey, expanded: !showReplies });
              }}
              replies={replies}
              threadId={id}
            />
          </Box>
        ) : null}
      </Box>
      <Box
        className="relative mt-2"
        hidden={!showReplies}
        id={`comment-replies-${id}`}
        ref={threadRef}
      >
        {showReplies ? (
          <>
            {hasReplies ? (
              <span
                aria-hidden
                className="border-border/70 pointer-events-none absolute -top-2 bottom-1 left-3.5 border-l"
              />
            ) : null}
            {subComments.map((reply) => (
              <CommentRow
                checkingMaya={checkingMaya}
                comment={reply}
                currentUserId={currentUserId}
                depth={depth + 1}
                highlightedId={highlightedId}
                isDialog={isDialog}
                key={reply.id}
                lastReadAt={lastReadAt}
                mayaPending={
                  !isSubComment
                    ? mayaPending.filter(({ run }) =>
                        flattenComments([reply]).some(
                          (source) => source.id === run.sourceId,
                        ),
                      )
                    : undefined
                }
                mayaRuns={mayaRuns}
                onCheckMaya={onCheckMaya}
                onRetryMaya={onRetryMaya}
                readObservedAt={readObservedAt}
                retryingMayaRunId={retryingMayaRunId}
                revealKey={revealKey}
                stateById={stateById}
                storyId={storyId}
                targetId={targetId}
                teamId={teamId}
                threadResolved={threadResolved || resolved}
              />
            ))}
            {pendingHere.map(({ run, status }) => (
              <CommentMayaPending
                checkAgainDisabled={checkingMaya}
                isDialog={isDialog}
                key={run.id}
                onCheckAgain={onCheckMaya}
                onRetry={
                  run.requesterId === currentUserId && onRetryMaya
                    ? () => {
                        onRetryMaya(run);
                      }
                    : undefined
                }
                retryDisabled={retryingMayaRunId === run.id}
                status={status}
              />
            ))}
          </>
        ) : null}
      </Box>
      {!isEditing ? (
        <CommentFooter
          commentId={id}
          disabled={disabled}
          onCollapse={
            canCollapse
              ? () => {
                  setExpansion({ revealKey, expanded: false });
                }
              : undefined
          }
          onReact={handleReact}
          onReply={
            canReply
              ? () => {
                  setExpansion({ revealKey, expanded: true });
                  setIsReplying(true);
                }
              : undefined
          }
          reactions={reactions}
        />
      ) : null}
      {isReplying ? (
        <Box className="mt-3 pr-1 pl-[2.4rem]">
          <CommentInput
            className={cn("mb-2 min-h-12 focus-within:shadow-none", {
              "dark:bg-background/40": isDialog,
            })}
            onCancel={handleCancel}
            parentId={id}
            storyId={storyId}
            teamId={teamId}
          />
        </Box>
      ) : null}
      {isEditing ? (
        <Box className="mt-3 pr-1 pl-8">
          <CommentInput
            className={cn("mb-2 min-h-12 focus-within:shadow-none", {
              "dark:bg-background/40": isDialog,
            })}
            commentId={id}
            initialComment={comment}
            onCancel={handleCancel}
            storyId={storyId}
            teamId={teamId}
          />
        </Box>
      ) : null}
      <CommentDeleteDialog
        onDelete={() => {
          setIsOpen(false);
          deleteComment({ commentId: id, storyId });
        }}
        onOpenChange={setIsOpen}
        open={isOpen}
      />
    </Box>
  );
};
