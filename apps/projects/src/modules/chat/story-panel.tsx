"use client";

import { lazy, Suspense, useEffect, useRef } from "react";
import { Button } from "ui";
import { useNativeLink } from "@/components/ui/link-preview/native-query";
import type { NativeLinkTarget } from "@/components/ui/link-preview/native-target";
import { useNativePreviewTerminology } from "@/components/ui/link-preview/native-terminology";
import { StoryCloseButton } from "@/modules/story/components/story-close-button";
import { CHAT_SIDE_PANEL_CLASS } from "./chat-styles";

const StoryPage = lazy(() =>
  import("@/modules/story").then((module) => ({ default: module.StoryPage })),
);

const StoryLoading = ({
  onClose,
  term,
}: {
  onClose: () => void;
  term: string;
}) => (
  <div className="flex items-center justify-between gap-3 px-5 py-3">
    <p className="text-text-muted" role="status">
      Loading {term}…
    </p>
    <StoryCloseButton onClose={onClose} term={term} />
  </div>
);

export const ChatStoryPanel = ({
  target,
  onClose,
}: {
  target: Extract<NativeLinkTarget, { kind: "story" }>;
  onClose: () => void;
}) => {
  const panelRef = useRef<HTMLElement>(null);
  const { data, loading, retry } = useNativeLink(target, true);
  const story = data?.kind === "story" ? data.value : null;
  const { label, term } = useNativePreviewTerminology(target, data);

  useEffect(() => {
    if (!panelRef.current?.contains(document.activeElement))
      panelRef.current?.focus();
  }, [target.href]);

  useEffect(() => {
    const handleEscape = (event: KeyboardEvent) => {
      if (
        event.key === "Escape" &&
        !event.defaultPrevented &&
        event.target instanceof Node &&
        panelRef.current?.contains(event.target)
      )
        onClose();
    };
    window.addEventListener("keydown", handleEscape);
    return () => {
      window.removeEventListener("keydown", handleEscape);
    };
  }, [onClose]);

  return (
    <aside
      aria-label={`${label} details`}
      className={CHAT_SIDE_PANEL_CLASS}
      ref={panelRef}
      tabIndex={-1}
    >
      <div className="notification-story-container min-h-0 flex-1 overflow-hidden">
        {loading ? <StoryLoading onClose={onClose} term={term} /> : null}
        {!loading && !story ? (
          <div className="space-y-3 p-5" role="alert">
            <div className="flex justify-end">
              <StoryCloseButton onClose={onClose} term={term} />
            </div>
            <p>This {term} is unavailable or you no longer have access.</p>
            <Button color="tertiary" onClick={retry} size="sm">
              Try again
            </Button>
          </div>
        ) : null}
        {story ? (
          <Suspense fallback={<StoryLoading onClose={onClose} term={term} />}>
            <StoryPage
              isNotifications
              key={story.id}
              onClose={onClose}
              storyId={story.id}
            />
          </Suspense>
        ) : null}
      </div>
    </aside>
  );
};
