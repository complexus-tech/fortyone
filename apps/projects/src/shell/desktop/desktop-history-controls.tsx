"use client";

import { Tooltip } from "ui";
import type { DesktopNavigationState } from "./bridge";

type DesktopHistoryControlsProps = DesktopNavigationState & {
  goBack: () => void;
  goForward: () => void;
};

const HistoryArrow = ({ forward = false }: { forward?: boolean }) => (
  <svg
    aria-hidden="true"
    className="size-[18px] shrink-0"
    fill="none"
    stroke="currentColor"
    strokeLinecap="round"
    strokeLinejoin="round"
    strokeWidth="1.6"
    viewBox="0 0 20 20"
  >
    <path d={forward ? "M3 10h14m-5-5 5 5-5 5" : "M17 10H3m5-5-5 5 5 5"} />
  </svg>
);

export const DesktopHistoryControls = ({
  canGoBack,
  canGoForward,
  goBack,
  goForward,
}: DesktopHistoryControlsProps) => {
  return (
    <div className="flex shrink-0 items-center gap-0.5" data-desktop-history>
      <Tooltip title="Back">
        <button
          aria-label="Back"
          className="text-foreground/80 hover:bg-state-hover focus-visible:ring-ring flex size-[32px] items-center justify-center rounded-md outline-none focus-visible:ring-2 disabled:pointer-events-none disabled:opacity-30"
          disabled={!canGoBack}
          onClick={goBack}
          type="button"
        >
          <HistoryArrow />
        </button>
      </Tooltip>
      <Tooltip title="Forward">
        <button
          aria-label="Forward"
          className="text-foreground/80 hover:bg-state-hover focus-visible:ring-ring flex size-[32px] items-center justify-center rounded-md outline-none focus-visible:ring-2 disabled:pointer-events-none disabled:opacity-30"
          disabled={!canGoForward}
          onClick={goForward}
          type="button"
        >
          <HistoryArrow forward />
        </button>
      </Tooltip>
    </div>
  );
};
