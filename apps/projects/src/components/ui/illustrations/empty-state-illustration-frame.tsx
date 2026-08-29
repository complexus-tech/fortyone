import type { ReactNode } from "react";
import { cn } from "lib";

export const EMPTY_STATE_SQUIRCLE_PATH =
  "M62 34H218C230.8 34 234 37.2 234 50V140C234 152.8 230.8 156 218 156H62C49.2 156 46 152.8 46 140V50C46 37.2 49.2 34 62 34Z";

export const EmptyStateIllustrationFrame = ({
  children,
  className,
}: {
  children: ReactNode;
  className?: string;
}) => (
  <svg
    aria-hidden="true"
    className={cn("text-primary h-auto w-64", className)}
    fill="none"
    viewBox="0 0 280 180"
    xmlns="http://www.w3.org/2000/svg"
  >
    <circle cx="140" cy="90" fill="currentColor" opacity="0.04" r="86" />
    <path d={EMPTY_STATE_SQUIRCLE_PATH} fill="currentColor" opacity="0.035" />
    <path
      d={EMPTY_STATE_SQUIRCLE_PATH}
      stroke="currentColor"
      strokeOpacity="0.24"
      strokeWidth="1.5"
    />
    {children}
  </svg>
);
