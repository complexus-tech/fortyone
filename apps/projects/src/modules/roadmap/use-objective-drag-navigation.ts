"use client";

import { useEffect, useRef } from "react";

// Dnd-kit keeps its document-capture click suppression for 50ms after release.
const DRAG_RELEASE_CLICK_WINDOW_MS = 50;

/** Prevent an anchor's native action even when Dnd-kit stops its React click handler. */
export const useObjectiveDragNavigation = (enabled: boolean) => {
  const draggedId = useRef<string | null>(null);
  const releaseTimer = useRef<ReturnType<typeof setTimeout> | null>(null);
  useEffect(() => {
    if (!enabled) return;
    const preventDragNavigation = (event: MouseEvent) => {
      if (!draggedId.current || !(event.target instanceof Element)) return;
      const link = event.target.closest<HTMLElement>("[data-objective-link]");
      if (link?.dataset.objectiveLink !== draggedId.current) return;
      event.preventDefault();
      draggedId.current = null;
    };
    window.addEventListener("click", preventDragNavigation, true);
    return () => {
      window.removeEventListener("click", preventDragNavigation, true);
      draggedId.current = null;
      if (releaseTimer.current !== null) clearTimeout(releaseTimer.current);
      releaseTimer.current = null;
    };
  }, [enabled]);
  return {
    start: (id: string) => {
      if (releaseTimer.current !== null) clearTimeout(releaseTimer.current);
      releaseTimer.current = null;
      draggedId.current = id;
    },
    release: () => {
      if (releaseTimer.current !== null) clearTimeout(releaseTimer.current);
      if (!draggedId.current) return;
      releaseTimer.current = setTimeout(() => {
        draggedId.current = null;
        releaseTimer.current = null;
      }, DRAG_RELEASE_CLICK_WINDOW_MS);
    },
  };
};
