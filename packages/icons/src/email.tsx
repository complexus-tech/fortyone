import { cn } from "lib";
import type { Icon } from "./types";

export const EmailIcon = (props: Icon) => {
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
        d="M7 8.5L9.94 10.24C11.66 11.25 12.34 11.25 14.06 10.24L17 8.5"
        strokeLinecap="round"
        strokeLinejoin="round"
      />
      <path
        d="M2.02 13.48C2.08 16.54 2.11 18.07 3.24 19.21C4.38 20.34 5.95 20.38 9.1 20.46C11.04 20.51 12.96 20.51 14.9 20.46C18.05 20.38 19.62 20.34 20.76 19.21C21.89 18.07 21.92 16.54 21.98 13.48C22.01 12.49 22.01 11.51 21.98 10.52C21.92 7.46 21.89 5.93 20.76 4.79C19.62 3.66 18.05 3.62 14.9 3.54C12.96 3.49 11.04 3.49 9.1 3.54C5.95 3.62 4.38 3.66 3.24 4.79C2.11 5.93 2.08 7.46 2.02 10.52C1.99 11.51 1.99 12.49 2.02 13.48Z"
        strokeLinejoin="round"
      />
    </svg>
  );
};
