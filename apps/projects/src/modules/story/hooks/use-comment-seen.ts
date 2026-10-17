import { useEffect, useRef } from "react";
import { useWorkspacePath } from "@/hooks/use-workspace-path";
import { markCommentSeen } from "../queries/comment-collaboration";

const VISIBLE_DWELL_MS = 500;
const MAX_ATTEMPTS = 3;

const observeCommentVisibility = (
  element: HTMLDivElement,
  acknowledge: () => Promise<void>,
) => {
  let active = true;
  let visible = false;
  let seen = false;
  let inFlight = false;
  let attempts = 0;
  let timer: ReturnType<typeof setTimeout> | undefined;
  const canAcknowledge = () =>
    active &&
    visible &&
    document.visibilityState === "visible" &&
    document.hasFocus();

  const update = () => {
    if (!canAcknowledge()) {
      clearTimeout(timer);
      timer = undefined;
      return;
    }
    if (seen || inFlight || timer || attempts >= MAX_ATTEMPTS) return;

    timer = setTimeout(
      () => {
        timer = undefined;
        if (!canAcknowledge()) return;
        inFlight = true;
        attempts++;
        void acknowledge().then(
          () => {
            seen = true;
            inFlight = false;
          },
          () => {
            inFlight = false;
            // Retry only while it is still visible. If receipts remain
            // unavailable, the server safely sends the notification.
            update();
          },
        );
      },
      attempts ? 1_000 : VISIBLE_DWELL_MS,
    );
  };

  const observer = new IntersectionObserver(
    ([entry]) => {
      visible = entry.isIntersecting && entry.intersectionRatio >= 0.5;
      update();
    },
    { threshold: 0.5 },
  );
  observer.observe(element);
  document.addEventListener("visibilitychange", update);
  window.addEventListener("focus", update);
  window.addEventListener("blur", update);

  return () => {
    active = false;
    clearTimeout(timer);
    observer.disconnect();
    document.removeEventListener("visibilitychange", update);
    window.removeEventListener("focus", update);
    window.removeEventListener("blur", update);
  };
};

/** Exact visibility receipts keep Maya's unseen replies eligible to notify. */
export const useCommentSeen = (
  storyId: string,
  commentId: string,
  enabled: boolean,
) => {
  const { workspaceSlug } = useWorkspacePath();
  const ref = useRef<HTMLDivElement>(null);

  useEffect(() => {
    const element = ref.current;
    if (!enabled || !element || typeof IntersectionObserver === "undefined")
      return;

    return observeCommentVisibility(element, () =>
      markCommentSeen(storyId, commentId, { workspaceSlug }),
    );
  }, [commentId, enabled, storyId, workspaceSlug]);

  return ref;
};
