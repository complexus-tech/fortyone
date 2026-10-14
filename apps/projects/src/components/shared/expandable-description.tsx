"use client";

import type { ReactNode } from "react";
import { useEffect, useId, useRef, useState } from "react";

const COLLAPSED_HEIGHT = 240;
const COLLAPSED_MASK =
  "linear-gradient(to bottom, black calc(100% - 6rem), rgb(0 0 0 / 0.92) calc(100% - 4.5rem), rgb(0 0 0 / 0.5) calc(100% - 3rem), rgb(0 0 0 / 0.08) calc(100% - 1.5rem), transparent 100%)";

export const ExpandableDescription = ({
  children,
}: {
  children: ReactNode;
}) => {
  const contentId = useId();
  const contentRef = useRef<HTMLDivElement>(null);
  const [canExpand, setCanExpand] = useState(false);
  const [isExpanded, setIsExpanded] = useState(false);

  useEffect(() => {
    const content = contentRef.current;
    if (!content) return;

    const measureOverflow = () => {
      setCanExpand(content.scrollHeight > COLLAPSED_HEIGHT + 1);
    };
    measureOverflow();

    // Observe the full content so media loads and responsive changes are measured
    // even while the outer preview has a fixed maximum height.
    const observer = new ResizeObserver(measureOverflow);
    observer.observe(content);
    return () => {
      observer.disconnect();
    };
  }, []);

  return (
    <div className="expandable-description">
      <div
        className={isExpanded ? undefined : "overflow-clip"}
        id={contentId}
        style={
          isExpanded
            ? undefined
            : {
                maxHeight: COLLAPSED_HEIGHT,
                maskImage: canExpand ? COLLAPSED_MASK : undefined,
              }
        }
      >
        <div
          className="flow-root"
          onFocusCapture={() => {
            // Never clip the caret, selected text, or keyboard-focused links.
            setIsExpanded(true);
          }}
          ref={contentRef}
        >
          {children}
        </div>
      </div>
      {canExpand ? (
        <button
          aria-controls={contentId}
          aria-expanded={isExpanded}
          className="text-foreground focus-visible:outline-ring mt-1 ml-2 block cursor-pointer rounded-sm py-1 underline-offset-4 hover:underline focus-visible:outline-2 focus-visible:outline-offset-2"
          onClick={(event) => {
            // Transfer focus explicitly so collapsing cannot leave an active caret
            // in hidden content, including browsers that do not focus clicked buttons.
            event.currentTarget.focus({ preventScroll: true });
            setIsExpanded((expanded) => !expanded);
          }}
          onMouseDown={(event) => {
            // Blur can hide the editor's trailing blank line and move this control.
            // Keep it still until onClick transfers focus and changes expansion.
            event.preventDefault();
          }}
          type="button"
        >
          {isExpanded ? "See less" : "See more"}
        </button>
      ) : null}
    </div>
  );
};
