import { cn } from "lib";
import type { Icon } from "./types";

// Hugeicons TargetDollarIcon, Stroke Rounded (MIT).
// Source: https://github.com/hugeicons/hugeicons/blob/9c48f3723dfb243909fa83501aa7c6423ab972c0/icons/target-dollar.svg
// Copyright (c) 2025 Hugeicons. See ../LICENSE.hugeicons.
export const TargetDollarIcon = (props: Icon) => {
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
      <path d="M22 12C22 6.47715 17.5228 2 12 2C6.47715 2 2 6.47715 2 12C2 17.5228 6.47715 22 12 22C17.5228 22 22 17.5228 22 12Z" />
      <path d="M12 2V4" strokeLinecap="round" />
      <path d="M22 12L20 12" strokeLinecap="round" />
      <path d="M12 20L12 22" strokeLinecap="round" />
      <path d="M4 12H2" strokeLinecap="round" />
      <path
        d="M12 9C10.8954 9 10 9.67157 10 10.5C10 11.3284 10.8954 12 12 12C13.1046 12 14 12.6716 14 13.5C14 14.3284 13.1046 15 12 15M12 9C12.8708 9 13.6116 9.4174 13.8862 10M12 9V8M12 15C11.1292 15 10.3884 14.5826 10.1138 14M12 15V16"
        strokeLinecap="round"
      />
    </svg>
  );
};
