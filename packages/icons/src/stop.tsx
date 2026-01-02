import type { Icon } from "./types";
import { cn } from "lib";

export const StopIcon = (props: Icon) => {
  const { className, strokeWidth = 2.5, ...rest } = props;
  return (
    <svg
      {...rest}
      className={cn(
        "h-[1.3rem] w-auto text-icon",
        className
      )}
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
        d="M3.25 7C3.25 4.92893 4.92893 3.25 7 3.25H17C19.0711 3.25 20.75 4.92893 20.75 7V17C20.75 19.0711 19.0711 20.75 17 20.75H7C4.92893 20.75 3.25 19.0711 3.25 17V7Z"
        fill="currentColor"
      />
    </svg>
  );
};
