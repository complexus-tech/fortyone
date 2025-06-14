import { cn } from "lib";
import type { Icon } from "./types";

export const GridIcon = (props: Icon) => {
  const { strokeWidth = 2, className, ...rest } = props;
  return (
    <svg
      {...rest}
      className={cn("h-5 w-auto text-gray dark:text-gray-300", className)}
      fill="currentColor"
      height="24"
      viewBox="0 0 24 24"
      width="24"
      xmlns="http://www.w3.org/2000/svg"
    >
      <path
        d="M16.5 22.5H18.5C20.7091 22.5 22.5 20.7091 22.5 18.5V5.5C22.5 3.29086 20.7091 1.5 18.5 1.5H16.5L16.5 22.5Z"
        fill="currentColor"
      />
      <path d="M14.5 1.5H9.50001L9.50037 22.5H14.5V1.5Z" fill="currentColor" />
      <path
        d="M5.5 1.5H7.50001L7.50037 22.5H5.5C3.29086 22.5 1.5 20.7091 1.5 18.5V5.5C1.5 3.29086 3.29086 1.5 5.5 1.5Z"
        fill="currentColor"
      />
    </svg>
  );
};
