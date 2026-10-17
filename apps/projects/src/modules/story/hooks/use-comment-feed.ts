import { useEffect, useMemo, useRef, useState } from "react";
import { useSearchParams } from "next/navigation";
import { useSession } from "@/lib/auth/client";
import { isStoryUuid } from "@/shared/routing/story";
import { normalizeCommentThreads } from "../lib/comment-tree";
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
  const comments = useMemo(() => {
    let threads = loadedComments;
    if (linkedThread) {
      threads = loadedComments.some((comment) => comment.id === linkedThread.id)
        ? loadedComments.map((comment) =>
            comment.id === linkedThread.id ? linkedThread : comment,
          )
        : [linkedThread, ...loadedComments];
    }
    return normalizeCommentThreads(threads);
  }, [loadedComments, linkedThread]);
  const newest = latestComment(comments);
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
  const focusedTarget = useRef({ key: revealKey, focused: false });
  const [highlight, setHighlight] = useState<{
    key: string;
    commentId: string;
  } | null>(null);

  useEffect(() => {
    if (focusedTarget.current.key !== revealKey)
      focusedTarget.current = { key: revealKey, focused: false };
    if (!targetId || !revealKey) return;
    // An off-page thread can arrive first. Wait for the initial feed so its
    // insertion does not move the focused reply out of view immediately after.
    if (commentsQuery.isPending) return;
    if (focusedTarget.current.focused) return;
    if (focusComment(targetId)) {
      focusedTarget.current.focused = true;
      setHighlight({ key: revealKey, commentId: targetId });
    }
  }, [
    targetId,
    revealKey,
    commentsQuery.data,
    commentsQuery.isPending,
    threadQuery.data,
  ]);

  useEffect(() => {
    if (!highlight || highlight.key !== revealKey) return;
    const timeout = window.setTimeout(() => {
      setHighlight((current) => (current === highlight ? null : current));
    }, 5_000);
    return () => {
      window.clearTimeout(timeout);
    };
  }, [highlight, revealKey]);

  const jumpToLatest = () => {
    const commentId = latestCommentId;
    if (!commentId) return;
    setJump((current) => ({
      requestedId,
      commentId,
      sequence: (current?.sequence ?? 0) + 1,
    }));
  };
  return {
    commentsQuery,
    stateQuery,
    mayaQuery,
    threadQuery,
    targetId,
    highlightedId: highlight?.key === revealKey ? highlight.commentId : null,
    revealKey,
    loadedTarget,
    comments,
    containerRef,
    lastReadAt,
    stateById,
    hasUnread,
    latestCommentId,
    jumpToLatest,
    currentUserId: session?.user.id,
  };
};
