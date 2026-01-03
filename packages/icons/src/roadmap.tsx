import { cn } from "lib";
import type { Icon } from "./types";

export const RoadmapIcon = (props: Icon) => {
  const { className, strokeWidth = 2, ...rest } = props;
  return (
    <svg
      {...rest}
      className={cn("h-5 w-auto text-icon", className)}
      fill="none"
      strokeWidth={strokeWidth}
      height="24"
      viewBox="0 0 24 24"
      width="24"
      xmlns="http://www.w3.org/2000/svg"
    >
      <path
        fillRule="evenodd"
        clipRule="evenodd"
        d="M7.25 2.53618C7.24992 2.53622 7.24984 2.53626 7.24976 2.5363L2.21738 5.05249C1.6245 5.34893 1.25 5.95489 1.25 6.61774V18.6123C1.25 19.8068 2.42022 20.6502 3.5534 20.2725L7.25 19.0403V2.53618ZM8.75 19.1371L14.25 21.4943V4.8623L8.75 2.50516V19.1371ZM15.75 5.2497V21.7497H21C21.9665 21.7497 22.75 20.9662 22.75 19.9997V6.9997C22.75 6.03321 21.9665 5.2497 21 5.2497H15.75Z"
        fill="currentColor"
      />
    </svg>
  );
};
