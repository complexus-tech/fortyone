import type { FeedbackItem } from "./types";

type Comment = NonNullable<FeedbackItem["comments"]>[number];
const byDate = (first: Comment, second: Comment) =>
  Date.parse(first.createdAt) - Date.parse(second.createdAt);

/** Keep each conversation together while retaining virtualized comment rows. */
export function threadedFeedbackComments(comments: Comment[]): Comment[] {
  const ids = new Set(comments.map((comment) => comment.id));
  const roots: Comment[] = [];
  const replies = new Map<string, Comment[]>();
  for (const comment of comments) {
    if (comment.parentId && ids.has(comment.parentId)) {
      const siblings = replies.get(comment.parentId) ?? [];
      siblings.push(comment);
      replies.set(comment.parentId, siblings);
    } else roots.push(comment);
  }
  return roots
    .sort((a, b) => byDate(b, a))
    .flatMap((comment) => [
      comment,
      ...(replies.get(comment.id) ?? []).sort(byDate),
    ]);
}
