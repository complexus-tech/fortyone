"use client";

import { cn } from "lib";
import { useEffect, useRef, useState } from "react";
import type { ReactNode } from "react";

const HIGHLIGHT_DURATION_MS = 2400;

export const resolveTargetKeyResultId = (
  requestedKeyResultId: string | null,
  keyResultIds: readonly string[],
) =>
  requestedKeyResultId && keyResultIds.includes(requestedKeyResultId)
    ? requestedKeyResultId
    : null;

export const KeyResultDeepLinkTarget = ({
  children,
  id,
  isTarget,
  name,
  viewport,
}: {
  children: ReactNode;
  id: string;
  isTarget: boolean;
  name: string;
  viewport: "desktop" | "mobile";
}) => {
  const rowRef = useRef<HTMLDivElement>(null);
  const [isHighlighted, setIsHighlighted] = useState(false);

  useEffect(() => {
    if (!isTarget) {
      setIsHighlighted(false);
      return;
    }

    const isDesktop = window.matchMedia("(min-width: 768px)").matches;
    const isActiveViewport = viewport === (isDesktop ? "desktop" : "mobile");
    if (!isActiveViewport) return;

    const row = rowRef.current;
    if (!row) return;

    setIsHighlighted(true);
    const frameId = window.requestAnimationFrame(() => {
      const prefersReducedMotion = window.matchMedia(
        "(prefers-reduced-motion: reduce)",
      ).matches;
      row.scrollIntoView({
        behavior: prefersReducedMotion ? "auto" : "smooth",
        block: "center",
        inline: "nearest",
      });
      row.focus({ preventScroll: true });
    });
    const timeoutId = window.setTimeout(() => {
      setIsHighlighted(false);
    }, HIGHLIGHT_DURATION_MS);

    return () => {
      window.cancelAnimationFrame(frameId);
      window.clearTimeout(timeoutId);
    };
  }, [isTarget, viewport]);

  return (
    <div
      aria-current={isTarget ? "location" : undefined}
      aria-label={`Key result: ${name}`}
      className={cn(
        "scroll-mt-24 rounded-2xl transition-[background-color,box-shadow] duration-700 outline-none motion-reduce:transition-none",
        isHighlighted && "bg-primary/5 ring-primary/30 ring-1",
      )}
      data-key-result-id={id}
      id={`key-result-${viewport}-${id}`}
      ref={rowRef}
      role="group"
      tabIndex={-1}
    >
      {children}
    </div>
  );
};
