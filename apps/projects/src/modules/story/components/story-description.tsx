import type { ReactNode } from "react";
import { useEffect, useId, useRef, useState } from "react";

const COLLAPSED_HEIGHT = 240;

export const StoryDescription = ({ children }: { children: ReactNode }) => {
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
    <div>
      <div
        className={isExpanded ? undefined : "overflow-clip"}
        id={contentId}
        style={
          isExpanded
            ? undefined
            : {
                maxHeight: COLLAPSED_HEIGHT,
                maskImage: canExpand
                  ? "linear-gradient(to bottom, black calc(100% - 3rem), transparent)"
                  : undefined,
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
          className="text-foreground focus-visible:outline-ring mt-2 ml-2 block cursor-pointer rounded-sm py-1 underline-offset-4 hover:underline focus-visible:outline-2 focus-visible:outline-offset-2"
          onClick={(event) => {
            // Safari does not focus buttons on pointer clicks. Move focus out of
            // the editor before collapsing so its caret cannot keep hidden text active.
            event.currentTarget.focus({ preventScroll: true });
            setIsExpanded((expanded) => !expanded);
          }}
          type="button"
        >
          {isExpanded ? "See less" : "See more"}
        </button>
      ) : null}
    </div>
  );
};
