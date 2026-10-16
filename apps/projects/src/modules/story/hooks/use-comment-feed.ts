import { useEffect, useMemo, useRef, useState } from "react";
import { useSearchParams } from "next/navigation";
import { useSession } from "@/lib/auth/client";
import { isStoryUuid } from "@/shared/routing/story";
import {
  findCommentThread,
  flattenComments,
  focusComment,
  isUnreadComment,
  latestComment,
} from "../lib/comment-navigation";
import { useStoryCommentsInfinite } from "./story-comments";
import {
  useCommentThread,
  useStoryCommentState,
} from "./comment-collaboration";
import { useCommentReadTracking } from "./comment-read-tracking";

export const useCommentFeed = (storyId: string) => {
  const { data: session } = useSession();
  const searchParams = useSearchParams();
  const requestedId = searchParams.get("comment");
  const [jump, setJump] = useState<{
    requestedId: string | null;
    commentId: string;
    sequence: number;
  }>();
  const requestedTarget =
    requestedId && isStoryUuid(requestedId) ? requestedId : null;
  const targetId =
    jump?.requestedId === requestedId ? jump.commentId : requestedTarget;
  const revealKey = targetId
    ? `${targetId}:${jump?.requestedId === requestedId ? jump.sequence : 0}`
    : null;
  const commentsQuery = useStoryCommentsInfinite(storyId);
  const stateQuery = useStoryCommentState(storyId);
  const loadedComments = useMemo(
    () =>
      commentsQuery.data?.pages
        .flatMap((page) => page.comments)
        .filter(
          (comment) => !comment.comment.includes("commented on GitHub issue #"),
        ) ?? [],
    [commentsQuery.data],
  );
  const loadedTarget = targetId
    ? findCommentThread(loadedComments, targetId)
    : undefined;
  const threadQuery = useCommentThread(storyId, loadedTarget ? null : targetId);
  const linkedThread = !threadQuery.isError ? threadQuery.data : undefined;
  let comments = loadedComments;
  if (linkedThread) {
    comments = loadedComments.some((comment) => comment.id === linkedThread.id)
      ? loadedComments.map((comment) =>
          comment.id === linkedThread.id ? linkedThread : comment,
        )
      : [linkedThread, ...loadedComments];
  }
  const newest = latestComment(comments);
  const { containerRef, lastReadAt, hadUnread } = useCommentReadTracking(
    storyId,
    stateQuery.data,
    newest?.createdAt,
  );
  const stateById = useMemo(
    () =>
      new Map(
        stateQuery.data?.comments.map((entry) => [entry.commentId, entry]) ??
          [],
      ),
    [stateQuery.data],
  );
  const hasLoadedUnread = flattenComments(comments).some((comment) =>
    isUnreadComment(comment, lastReadAt, session?.user.id),
  );
  const hasUnread =
    hasLoadedUnread || hadUnread || (stateQuery.data?.unreadCount ?? 0) > 0;
  const focusedTarget = useRef<string | null>(null);

  useEffect(() => {
    if (!targetId) {
      focusedTarget.current = null;
      return;
    }
    if (focusedTarget.current === revealKey) return;
    if (focusComment(targetId)) focusedTarget.current = revealKey;
  }, [targetId, revealKey, commentsQuery.data, threadQuery.data]);

  const jumpToLatest = () => {
    const commentId = stateQuery.data?.latestCommentId ?? newest?.id;
    if (!commentId) return;
    setJump((current) => ({
      requestedId,
      commentId,
      sequence: (current?.sequence ?? 0) + 1,
    }));
    focusComment(commentId);
  };
  return {
    commentsQuery,
    stateQuery,
    threadQuery,
    targetId,
    revealKey,
    loadedTarget,
    comments,
    containerRef,
    lastReadAt,
    stateById,
    hasUnread,
    jumpToLatest,
    currentUserId: session?.user.id,
  };
};
