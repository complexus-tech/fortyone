import type { ReactNode } from "react";
import { useEffect, useId, useRef, useState } from "react";
import { Button } from "ui";

const COLLAPSED_HEIGHT = 320;

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
    <div className="mb-10">
      <div
        className={isExpanded ? undefined : "overflow-hidden"}
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
        <Button
          aria-controls={contentId}
          aria-expanded={isExpanded}
          className="text-text-muted mt-2 -ml-2"
          color="tertiary"
          onClick={() => {
            setIsExpanded((expanded) => !expanded);
          }}
          size="sm"
          type="button"
          variant="naked"
        >
          {isExpanded ? "See less" : "See more"}
        </Button>
      ) : null}
    </div>
  );
};
