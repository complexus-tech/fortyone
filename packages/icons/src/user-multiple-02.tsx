import { cn } from "lib";
import type { Icon } from "./types";

export const UserMultiple02Icon = (props: Icon) => {
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
        d="M16.5 20V17.9704C16.5 16.7281 15.9407 15.5099 14.8103 14.9946C13.4315 14.3661 11.7779 14 10 14C8.22212 14 6.5685 14.3661 5.18968 14.9946C4.05927 15.5099 3.5 16.7281 3.5 17.9704V20"
        strokeLinecap="round"
        strokeLinejoin="round"
      />
      <path
        d="M20.5 20.001V17.9713C20.5 16.729 19.9407 15.5109 18.8103 14.9956C18.5497 14.8768 18.2792 14.7673 18 14.668"
        strokeLinecap="round"
        strokeLinejoin="round"
      />
      <circle cx="10" cy="7.5" r="3.5" />
      <path
        d="M15 4.14453C16.4457 4.57481 17.5 5.91408 17.5 7.49959C17.5 9.0851 16.4457 10.4244 15 10.8547"
        strokeLinecap="round"
        strokeLinejoin="round"
      />
    </svg>
  );
};
