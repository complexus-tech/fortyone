import { cn } from "lib";
import type { Icon } from "./types";

export const InviteMembersIcon = (props: Icon) => {
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
        d="M3 20V17.9704C3 16.7281 3.55927 15.5099 4.68968 14.9946C6.0685 14.3661 7.72212 14 9.5 14C10.7448 14 11.9287 14.1795 13 14.5028"
        strokeLinecap="round"
        strokeLinejoin="round"
      />
      <circle cx="9.5" cy="7.5" r="3.5" />
      <path
        d="M14.5 4.14453C15.9457 4.57481 17 5.91408 17 7.49959C17 9.0851 15.9457 10.4244 14.5 10.8547"
        strokeLinecap="round"
        strokeLinejoin="round"
      />
      <path
        d="M18 14V20M15 17H21"
        strokeLinecap="round"
        strokeLinejoin="round"
      />
    </svg>
  );
};
