import { cn } from "lib";
import type { Icon } from "./types";

// Hugeicons Calculator01Icon, Stroke Rounded (MIT).
// Source: https://github.com/hugeicons/hugeicons/blob/9c48f3723dfb243909fa83501aa7c6423ab972c0/icons/calculator-01.svg
// Copyright (c) 2025 Hugeicons. See ../LICENSE.hugeicons.
export const Calculator01Icon = (props: Icon) => {
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
      <path d="M3 10H21" strokeLinejoin="round" />
      <path d="M15 6L17 6" strokeLinecap="round" strokeLinejoin="round" />
      <path d="M21 13V11C21 6.75736 21 4.63604 19.682 3.31802C18.364 2 16.2426 2 12 2C7.75736 2 5.63604 2 4.31802 3.31802C3 4.63604 3 6.75736 3 11V13C3 17.2426 3 19.364 4.31802 20.682C5.63604 22 7.75736 22 12 22C16.2426 22 18.364 22 19.682 20.682C21 19.364 21 17.2426 21 13Z" />
      <path
        d="M7 14H7.52632M11.7368 14H12.2632M16.4737 14H17"
        strokeLinecap="round"
        strokeLinejoin="round"
      />
      <path
        d="M7 18H7.52632M11.7368 18H12.2632M16.4737 18H17"
        strokeLinecap="round"
        strokeLinejoin="round"
      />
    </svg>
  );
};
