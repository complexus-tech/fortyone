import { useEffect, useRef, useState } from "react";
import type { StoryCommentState } from "../types/comment-collaboration";
import { useMarkCommentsRead } from "./comment-collaboration";

export const useCommentReadTracking = (
  storyId: string,
  state: StoryCommentState | undefined,
  latestLoadedAt: string | undefined,
) => {
  const containerRef = useRef<HTMLDivElement>(null);
  const [visit, setVisit] = useState<{
    lastReadAt: string | null;
    hasUnread: boolean;
  }>();
  const lastReadAt = visit?.lastReadAt;
  const [visible, setVisible] = useState(false);
  const attemptedAt = useRef<string | undefined>(undefined);
  const { mutate } = useMarkCommentsRead(storyId);

  // Capture once when the first snapshot arrives. The keyed feed resets this
  // snapshot for another account/story; later durable markers cannot erase it.
  if (!visit && state) {
    setVisit({
      lastReadAt: state.lastReadAt,
      hasUnread: state.unreadCount > 0,
    });
  }

  useEffect(() => {
    const container = containerRef.current;
    if (!container) return;
    let intersects = false;
    const update = () => {
      setVisible(
        intersects &&
          document.visibilityState === "visible" &&
          document.hasFocus(),
      );
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
  }, []);

  useEffect(() => {
    if (!visible || !state || !latestLoadedAt || lastReadAt === undefined)
      return;
    // Do not consume comments newer than the rendered list or server snapshot.
    const cutoff =
      Date.parse(latestLoadedAt) < Date.parse(state.observedAt)
        ? latestLoadedAt
        : state.observedAt;
    if (state.lastReadAt && Date.parse(state.lastReadAt) >= Date.parse(cutoff))
      return;
    if (attemptedAt.current === cutoff) return;
    attemptedAt.current = cutoff;
    mutate(cutoff, {
      onError: () => {
        attemptedAt.current = undefined;
      },
    });
  }, [lastReadAt, latestLoadedAt, mutate, state, visible]);

  return { containerRef, lastReadAt, hadUnread: visit?.hasUnread ?? false };
};
