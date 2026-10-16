import { useEffect, useRef } from "react";
import type { StoryCommentState } from "../types/comment-collaboration";
import { useMarkCommentsRead } from "./comment-collaboration";

const useVisibleCommentRead = (
  storyId: string,
  {
    enabled = true,
    latestLoadedAt,
    observedAt,
    lastReadAt,
    threadId,
  }: {
    enabled?: boolean;
    latestLoadedAt?: string;
    observedAt?: string;
    lastReadAt?: string | null;
    threadId?: string;
  },
) => {
  const containerRef = useRef<HTMLDivElement>(null);
  const attemptedAt = useRef<string | undefined>(undefined);
  const { mutate } = useMarkCommentsRead(storyId);
  let cutoff: string | undefined;
  if (latestLoadedAt && observedAt) {
    cutoff = Date.parse(latestLoadedAt) < Date.parse(observedAt)
      ? latestLoadedAt
      : observedAt;
  }

  useEffect(() => {
    const container = containerRef.current;
    if (!container || !enabled || !cutoff) return;
    if (lastReadAt && Date.parse(lastReadAt) >= Date.parse(cutoff)) return;
    let intersects = false;
    const update = () => {
      if (
        !intersects ||
        document.visibilityState !== "visible" ||
        !document.hasFocus()
      )
        return;
      const attempt = `${cutoff}:${observedAt}`;
      if (attemptedAt.current === attempt) return;
      attemptedAt.current = attempt;
      // A failed automatic receipt waits for the next server snapshot, rather
      // than retrying immediately when its optimistic marker is rolled back.
      mutate({ observedAt: cutoff, ...(threadId ? { threadId } : {}) });
    };
    const observer = new IntersectionObserver(([entry]) => {
      intersects = entry.isIntersecting;
      update();
    });
    observer.observe(container);
    document.addEventListener("visibilitychange", update);
    window.addEventListener("focus", update);
    window.addEventListener("blur", update);
    return () => {
      observer.disconnect();
      document.removeEventListener("visibilitychange", update);
      window.removeEventListener("focus", update);
      window.removeEventListener("blur", update);
    };
  }, [cutoff, enabled, lastReadAt, mutate, observedAt, threadId]);

  return containerRef;
};

export const useCommentReadTracking = (
  storyId: string,
  state: StoryCommentState | undefined,
  latestRootLoadedAt: string | undefined,
) => {
  const containerRef = useVisibleCommentRead(storyId, {
    latestLoadedAt: latestRootLoadedAt,
    observedAt: state?.observedAt,
    lastReadAt: state?.lastReadAt,
  });
  return { containerRef, lastReadAt: state?.lastReadAt };
};

export const useCommentThreadReadTracking = (
  storyId: string,
  threadId: string,
  expanded: boolean,
  latestReplyAt: string | undefined,
  observedAt: string | undefined,
  lastReadAt: string | null | undefined,
) =>
  useVisibleCommentRead(storyId, {
    enabled: expanded,
    threadId,
    latestLoadedAt: latestReplyAt,
    observedAt,
    lastReadAt,
  });
