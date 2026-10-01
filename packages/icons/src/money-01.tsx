import { cn } from "lib";
import type { Icon } from "./types";

// Hugeicons Money01Icon, Stroke Rounded (MIT).
// Source: https://github.com/hugeicons/hugeicons/blob/9c48f3723dfb243909fa83501aa7c6423ab972c0/icons/money-01.svg
// Copyright (c) 2025 Hugeicons. See ../LICENSE.hugeicons.
export const Money01Icon = (props: Icon) => {
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
      <path
        d="M2.01733 15C4.2169 15 6.00001 16.7831 6.00001 18.9827"
        strokeLinecap="round"
        strokeLinejoin="round"
      />
      <path
        d="M18 18.9827V18.8908C18 16.742 19.742 15 21.8908 15"
        strokeLinecap="round"
        strokeLinejoin="round"
      />
      <path
        d="M6.00001 5.01733C6.00001 7.2169 4.2169 9.00001 2.01733 9.00001"
        strokeLinecap="round"
        strokeLinejoin="round"
      />
      <path
        d="M18 5.01733C18 7.19765 19.769 8.96876 21.9423 8.9996"
        strokeLinecap="round"
        strokeLinejoin="round"
      />
      <path
        d="M16 5H8C5.17157 5 3.75736 5 2.87868 5.87868C2 6.75736 2 8.17157 2 11V13C2 15.8284 2 17.2426 2.87868 18.1213C3.75736 19 5.17157 19 8 19H16C18.8284 19 20.2426 19 21.1213 18.1213C22 17.2426 22 15.8284 22 13V11C22 8.17157 22 6.75736 21.1213 5.87868C20.2426 5 18.8284 5 16 5Z"
        strokeLinecap="round"
        strokeLinejoin="round"
      />
      <path
        d="M15 12C15 13.6569 13.6569 15 12 15C10.3431 15 9 13.6569 9 12C9 10.3431 10.3431 9 12 9C13.6569 9 15 10.3431 15 12Z"
        strokeLinecap="round"
        strokeLinejoin="round"
      />
    </svg>
  );
};
