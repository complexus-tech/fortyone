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
import { useMayaRunRefresh, useStoryMayaRuns } from "./comment-maya";

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
  const mayaQuery = useStoryMayaRuns(storyId, { poll: true });
  useMayaRunRefresh(storyId, mayaQuery.data?.runs);
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
  const defaultExpandedThreadId = comments.reduce<
    { id: string; at: string } | undefined
  >((latest, comment) => {
    const reply = latestComment(comment.subComments);
    return reply &&
      (!latest || Date.parse(reply.createdAt) > Date.parse(latest.at))
      ? { id: comment.id, at: reply.createdAt }
      : latest;
  }, undefined)?.id;
  const latestRootLoadedAt = comments.reduce<string | undefined>(
    (latest, comment) => {
      if (comment.id.startsWith("pending-comment:")) return latest;
      return !latest || Date.parse(comment.createdAt) > Date.parse(latest)
        ? comment.createdAt
        : latest;
    },
    undefined,
  );
  const { containerRef, lastReadAt } = useCommentReadTracking(
    storyId,
    stateQuery.data,
    latestRootLoadedAt,
  );
  const stateById = useMemo(
    () =>
      new Map(
        stateQuery.data?.comments.map((entry) => [entry.commentId, entry]) ??
          [],
      ),
    [stateQuery.data],
  );
  const hasLoadedUnread = comments.some(
    (comment) =>
      isUnreadComment(comment, lastReadAt, session?.user.id) ||
      flattenComments(comment.subComments).some((reply) =>
        isUnreadComment(
          reply,
          stateQuery.data
            ? stateQuery.data.threadReads?.[comment.id] ?? null
            : undefined,
          session?.user.id,
        ),
      ),
  );
  const hasUnread = hasLoadedUnread || (stateQuery.data?.unreadCount ?? 0) > 0;
  const latestCommentId = stateQuery.data?.latestCommentId ?? newest?.id;
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
    const commentId = latestCommentId;
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
    mayaQuery,
    threadQuery,
    targetId,
    revealKey,
    loadedTarget,
    comments,
    containerRef,
    lastReadAt,
    stateById,
    hasUnread,
    defaultExpandedThreadId,
    latestCommentId,
    jumpToLatest,
    currentUserId: session?.user.id,
  };
};
