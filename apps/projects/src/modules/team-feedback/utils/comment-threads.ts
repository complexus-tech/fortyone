import type { TeamFeedbackComment } from "../types";

export type TeamFeedbackCommentThread = {
  comment: TeamFeedbackComment;
  replies: TeamFeedbackComment[];
};

const compareCommentDates = (
  first: TeamFeedbackComment,
  second: TeamFeedbackComment,
) => new Date(first.createdAt).getTime() - new Date(second.createdAt).getTime();

export const getCommentThreads = (
  comments: TeamFeedbackComment[],
): TeamFeedbackCommentThread[] => {
  const repliesByParent = new Map<string, TeamFeedbackComment[]>();
  const topLevelComments: TeamFeedbackComment[] = [];

  for (const comment of comments) {
    if (comment.parentId) {
      const replies = repliesByParent.get(comment.parentId) ?? [];
      replies.push(comment);
      repliesByParent.set(comment.parentId, replies);
    } else {
      topLevelComments.push(comment);
    }
  }

  return topLevelComments
    .sort((first, second) => compareCommentDates(second, first))
    .map((comment) => ({
      comment,
      replies: (repliesByParent.get(comment.id) ?? []).sort(
        compareCommentDates,
      ),
    }));
};
