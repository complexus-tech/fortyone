import { cn } from "lib";
import type { Icon } from "./types";

// Hugeicons Coins01Icon, Stroke Rounded (MIT).
// Source: https://github.com/hugeicons/hugeicons/blob/9c48f3723dfb243909fa83501aa7c6423ab972c0/icons/coins-01.svg
// Copyright (c) 2025 Hugeicons. See ../LICENSE.hugeicons.
export const Coins01Icon = (props: Icon) => {
  const { className, strokeWidth = 2, ...rest } = props;

  return (
    <svg
      {...rest}
      className={cn("h-5 w-auto text-icon", className)}
      fill="none"
      height="24"
      stroke="currentColor"
      strokeWidth={strokeWidth}
      viewBox="0 0 24 24"
      width="24"
      xmlns="http://www.w3.org/2000/svg"
    >
      <ellipse cx="15.5" cy="11" rx="6.5" ry="2" />
      <path d="M22 15.5C22 16.6046 19.0899 17.5 15.5 17.5C11.9101 17.5 9 16.6046 9 15.5" />
      <path d="M22 11V19.8C22 21.015 19.0899 22 15.5 22C11.9101 22 9 21.015 9 19.8V11" />
      <ellipse cx="8.5" cy="4" rx="6.5" ry="2" />
      <path
        d="M6 11C4.10819 10.7698 2.36991 10.1745 2 9M6 16C4.10819 15.7698 2.36991 15.1745 2 14"
        strokeLinecap="round"
      />
      <path
        d="M6 21C4.10819 20.7698 2.36991 20.1745 2 19L2 4"
        strokeLinecap="round"
      />
      <path d="M15 6V4" strokeLinecap="round" />
    </svg>
  );
};
