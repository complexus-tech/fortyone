import type { Icon } from "./types";
import { cn } from "lib";

export const VoiceIcon = (props: Icon) => {
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
      <path d="M3.5 10.5V13.5" strokeLinecap="round" />
      <path d="M7.75 8V16" strokeLinecap="round" />
      <path d="M12 5V19" strokeLinecap="round" />
      <path d="M16.25 8V16" strokeLinecap="round" />
      <path d="M20.5 10.5V13.5" strokeLinecap="round" />
    </svg>
  );
};
