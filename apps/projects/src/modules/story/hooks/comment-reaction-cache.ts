import type {
  CommentEmoji,
  CommentReaction,
  StoryCommentState,
} from "../types/comment-collaboration";

export const setCachedCommentReaction = (
  current: StoryCommentState | undefined,
  commentId: string,
  emoji: CommentEmoji,
  reaction: CommentReaction | undefined,
  removeEmptyComment = false,
): StoryCommentState | undefined => {
  if (!current) return current;
  const previous = current.comments.find(
    (comment) => comment.commentId === commentId,
  );
  const reactions = previous?.reactions ?? [];
  let nextReactions: CommentReaction[];
  if (!reaction) {
    nextReactions = reactions.filter((item) => item.emoji !== emoji);
  } else if (reactions.some((item) => item.emoji === emoji)) {
    nextReactions = reactions.map((item) =>
      item.emoji === emoji ? reaction : item,
    );
  } else {
    nextReactions = [...reactions, reaction];
  }
  const next = {
    commentId,
    resolvedAt: null,
    resolvedBy: null,
    ...previous,
    reactions: nextReactions,
  };

  if (removeEmptyComment && !next.resolvedAt && !next.reactions.length) {
    return {
      ...current,
      comments: current.comments.filter(
        (comment) => comment.commentId !== commentId,
      ),
    };
  }
  if (!previous && !nextReactions.length) return current;

  return {
    ...current,
    comments: previous
      ? current.comments.map((comment) =>
          comment.commentId === commentId ? next : comment,
        )
      : [...current.comments, next],
  };
};
