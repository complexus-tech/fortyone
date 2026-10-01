import { cn } from "lib";
import type { Icon } from "./types";

// Hugeicons Wallet01Icon, Stroke Rounded (MIT).
// Source: https://github.com/hugeicons/hugeicons/blob/9c48f3723dfb243909fa83501aa7c6423ab972c0/icons/wallet-01.svg
// Copyright (c) 2025 Hugeicons. See ../LICENSE.hugeicons.
export const Wallet01Icon = (props: Icon) => {
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
        d="M14 3H5C3.89543 3 3 3.89543 3 5C3 6.10457 3.89543 7 5 7H18C18 6.07003 18 5.60504 17.8978 5.22354C17.6204 4.18827 16.8117 3.37962 15.7765 3.10222C15.395 3 14.93 3 14 3Z"
        strokeLinecap="round"
        strokeLinejoin="round"
      />
      <path
        d="M3 5V15C3 17.8284 3 19.2426 3.87868 20.1213C4.75736 21 6.17157 21 9 21H15C17.8284 21 19.2426 21 20.1213 20.1213C21 19.2426 21 17.8284 21 15V13C21 10.1716 21 8.75736 20.1213 7.87868C19.2426 7 17.8284 7 15 7H7"
        strokeLinecap="round"
        strokeLinejoin="round"
      />
      <path
        d="M21 12H19C18.535 12 18.3025 12 18.1118 12.0511C17.5941 12.1898 17.1898 12.5941 17.0511 13.1118C17 13.3025 17 13.535 17 14C17 14.465 17 14.6975 17.0511 14.8882C17.1898 15.4059 17.5941 15.8102 18.1118 15.9489C18.3025 16 18.535 16 19 16H21"
        strokeLinecap="round"
        strokeLinejoin="round"
      />
    </svg>
  );
};
