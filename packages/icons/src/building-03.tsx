import { cn } from "lib";
import type { Icon } from "./types";

// Hugeicons Building03Icon, Stroke Rounded (MIT).
// Source: https://github.com/hugeicons/hugeicons/blob/9c48f3723dfb243909fa83501aa7c6423ab972c0/icons/building-03.svg
// Copyright (c) 2025 Hugeicons. See ../LICENSE.hugeicons.
export const Building03Icon = (props: Icon) => {
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
        d="M16 10L18.1494 10.6448C19.5226 11.0568 20.2092 11.2628 20.6046 11.7942C21 12.3256 21 13.0425 21 14.4761V22"
        strokeLinejoin="round"
      />
      <path
        d="M8 9L11 9M8 13L11 13"
        strokeLinecap="round"
        strokeLinejoin="round"
      />
      <path
        d="M12 22V19C12 18.0572 12 17.5858 11.7071 17.2929C11.4142 17 10.9428 17 10 17H9C8.05719 17 7.58579 17 7.29289 17.2929C7 17.5858 7 18.0572 7 19V22"
        strokeLinejoin="round"
      />
      <path d="M2 22L22 22" strokeLinecap="round" />
      <path
        d="M3 22V6.71724C3 4.20649 3 2.95111 3.79118 2.32824C4.58237 1.70537 5.74742 2.04355 8.07752 2.7199L13.0775 4.17122C14.4836 4.57937 15.1867 4.78344 15.5933 5.33965C16 5.89587 16 6.65344 16 8.16857V22"
        strokeLinecap="round"
        strokeLinejoin="round"
      />
    </svg>
  );
};
