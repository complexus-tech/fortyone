import type { Comment } from "@/types/comment";

export const flattenComments = (comments: Comment[]): Comment[] =>
  comments.flatMap((comment) => [
    comment,
    ...flattenComments(comment.subComments),
  ]);

export const findCommentThread = (comments: Comment[], commentId: string) =>
  comments.find((comment) =>
    flattenComments([comment]).some((item) => item.id === commentId),
  );

export const isUnreadComment = (
  comment: Comment,
  lastReadAt: string | null | undefined,
  userId?: string,
) =>
  lastReadAt !== undefined &&
  comment.userId !== userId &&
  !comment.id.startsWith("pending-comment:") &&
  (lastReadAt === null ||
    Date.parse(comment.createdAt) > Date.parse(lastReadAt));

export const latestComment = (comments: Comment[]) =>
  flattenComments(comments)
    .filter((comment) => !comment.id.startsWith("pending-comment:"))
    .reduce<Comment | undefined>(
      (latest, comment) =>
        !latest || Date.parse(comment.createdAt) > Date.parse(latest.createdAt)
          ? comment
          : latest,
      undefined,
    );

export const focusComment = (commentId: string) => {
  const element = document.getElementById(`comment-${commentId}`);
  if (!element) return false;
  element.scrollIntoView({ block: "center", behavior: "instant" });
  element.focus({ preventScroll: true });
  return true;
};
