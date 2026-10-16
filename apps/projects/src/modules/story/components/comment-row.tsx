import { useState } from "react";
import { Box, Button, Flex, Text, TimeAgo } from "ui";
import { CheckIcon, ReplyIcon } from "icons";
import { cn } from "lib";
import type { Comment } from "@/types/comment";
import { useDeleteCommentMutation } from "@/lib/hooks/delete-comment-mutation";
import { useCommentCollaboration } from "../hooks/comment-collaboration";
import { flattenComments, isUnreadComment } from "../lib/comment-navigation";
import type { CommentState } from "../types/comment-collaboration";
import { CommentInput } from "./comment-input";
import { CommentAuthor } from "./comment-author";
import { CommentActions } from "./comment-actions";
import { CommentReactions } from "./comment-reactions";
import { CommentBody } from "./comment-body";
import { CommentReplySummary } from "./comment-reply-summary";
import { CommentDeleteDialog } from "./comment-delete-dialog";

type CommentRowProps = {
  comment: Comment;
  storyId: string;
  teamId: string;
  isDialog?: boolean;
  isSubComment?: boolean;
  currentUserId?: string;
  stateById: ReadonlyMap<string, CommentState>;
  targetId: string | null;
  revealKey?: string | null;
  lastReadAt: string | null | undefined;
  showUnreadDivider?: boolean;
};

export const CommentRow = ({
  comment: { id, userId, user, createdAt, updatedAt, comment, subComments },
  storyId,
  teamId,
  isDialog,
  isSubComment = false,
  currentUserId,
  stateById,
  targetId,
  revealKey = targetId,
  lastReadAt,
  showUnreadDivider,
}: CommentRowProps) => {
  const { mutate: deleteComment, isPending: isDeleting } =
    useDeleteCommentMutation();
  const { reaction, resolution } = useCommentCollaboration(storyId, id);
  const [isOpen, setIsOpen] = useState(false);
  const [isReplying, setIsReplying] = useState(false);
  const [isEditing, setIsEditing] = useState(false);
  const [expansion, setExpansion] = useState({
    revealKey: null as string | null,
    expanded: false,
  });
  const replies = flattenComments(subComments);
  const hasTargetReply = replies.some((reply) => reply.id === targetId);
  const showReplies =
    expansion.revealKey === revealKey ? expansion.expanded : hasTargetReply;
  const isPending = id.startsWith("pending-comment:");
  const state = stateById.get(id);
  const resolved = Boolean(state?.resolvedAt);
  const reactions = state?.reactions ?? [];
  const disabled =
    isPending || isDeleting || reaction.isPending || resolution.isPending;
  const handleCancel = () => {
    setIsReplying(false);
    setIsEditing(false);
  };
  const handleReact = (
    emoji: Parameters<typeof reaction.mutate>[0]["emoji"],
    reacted: boolean,
  ) => {
    reaction.mutate({ emoji, reacted });
  };

  return (
    <Box
      className={cn("relative scroll-mt-24 pb-3 outline-none", {
        "ml-9 pt-1 pb-0": isSubComment,
        "bg-primary/5 ring-primary/20 rounded-md ring-1": targetId === id,
      })}
      id={`comment-${id}`}
      tabIndex={-1}
    >
      {showUnreadDivider ? (
        <Flex align="center" className="mb-3 ml-10" gap={2}>
          <Text className="shrink-0 text-xs" color="primary">
            New comments
          </Text>
          <div className="bg-primary/20 h-px flex-1" />
        </Flex>
      ) : null}
      <Box className="relative">
        {replies.length > 0 ? (
          <span
            aria-hidden
            className="border-border/70 pointer-events-none absolute top-8 bottom-4 left-3.5 w-[1.35rem] rounded-bl-lg border-b border-l"
          />
        ) : null}
        <Flex align="center" className="group flex-wrap" gap={1}>
          <CommentAuthor isDialog={isDialog} user={user} userId={userId} />
          <Text className="mx-0.5 text-[0.95rem]" color="muted">
            ·
          </Text>
          <Text className="text-[0.95rem]" color="muted">
            <TimeAgo timestamp={createdAt} />
          </Text>
          {Date.parse(updatedAt) > Date.parse(createdAt) ? (
            <Text
              className="ml-1 text-xs"
              color="muted"
              title={`Edited ${new Date(updatedAt).toUTCString()}`}
            >
              (edited)
            </Text>
          ) : null}
          {resolved ? (
            <Text
              className="ml-2 flex items-center gap-1 text-xs"
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
            onReact={handleReact}
            onResolve={() => {
              resolution.mutate(!resolved);
            }}
            reactions={reactions}
            resolved={resolved}
            storyId={storyId}
          />
        </Flex>
        {!isEditing ? (
          <>
            <CommentBody className="mt-0.5 ml-9" html={comment} />
            <CommentReactions
              disabled={disabled}
              onReact={handleReact}
              reactions={reactions}
            />
          </>
        ) : null}
        {replies.length > 0 ? (
          <Box className="mt-2">
            <CommentReplySummary
              expanded={showReplies}
              hasUnreadReplies={replies.some((reply) =>
                isUnreadComment(reply, lastReadAt, currentUserId),
              )}
              isDialog={isDialog}
              onToggle={() => {
                setExpansion({ revealKey, expanded: !showReplies });
              }}
              replies={replies}
              threadId={id}
            />
          </Box>
        ) : null}
      </Box>
      <Box className="mt-2" hidden={!showReplies} id={`comment-replies-${id}`}>
        {showReplies ? (
          <>
            {subComments.map((reply, index) => (
              <CommentRow
                comment={reply}
                currentUserId={currentUserId}
                isDialog={isDialog}
                isSubComment
                key={reply.id}
                lastReadAt={lastReadAt}
                revealKey={revealKey}
                showUnreadDivider={
                  isUnreadComment(reply, lastReadAt, currentUserId) &&
                  (index === 0 ||
                    !isUnreadComment(
                      subComments[index - 1],
                      lastReadAt,
                      currentUserId,
                    ))
                }
                stateById={stateById}
                storyId={storyId}
                targetId={targetId}
                teamId={teamId}
              />
            ))}
          </>
        ) : null}
      </Box>
      {!isSubComment && !isReplying && !isEditing && !resolved ? (
        <Button
          className="mt-3 ml-10 px-2"
          color="tertiary"
          disabled={disabled}
          leftIcon={<ReplyIcon className="h-4" />}
          onClick={() => {
            setExpansion({ revealKey, expanded: true });
            setIsReplying(true);
          }}
          size="sm"
        >
          Reply
        </Button>
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
        isSubComment={isSubComment}
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
