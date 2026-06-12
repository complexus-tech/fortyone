import { cn } from "lib";
import type { Icon } from "./types";

export const RequestsIcon = (props: Icon) => {
  const { className, strokeWidth = 2.2, ...rest } = props;
  return (
    <svg
      {...rest}
      className={cn("h-5 w-auto text-icon", className)}
      fill="none"
      height="24"
      stroke="currentColor"
      strokeLinecap="round"
      strokeLinejoin="round"
      strokeWidth={strokeWidth}
      viewBox="0 0 24 24"
      width="24"
      xmlns="http://www.w3.org/2000/svg"
    >
      <path d="M12.75 2.75H7.03068C4.16309 2.75 2.72929 2.75 2.33674 3.63757C1.94419 4.52514 2.90335 5.59831 4.82167 7.74467L8.7648 12.1565C9.01618 12.4378 9.14189 12.5784 9.20764 12.751C9.2734 12.9235 9.2734 13.1128 9.2734 13.4912V18.3704C9.2734 20.0513 9.2734 20.8917 9.80655 21.1712C10.3397 21.4507 11.0234 20.9688 12.3909 20.0049L13.3815 19.3066C13.7956 19.0147 14.0027 18.8688 14.1146 18.6522C14.2266 18.4357 14.2266 18.1812 14.2266 17.6722V13.4912C14.2266 13.1128 14.2266 12.9235 14.2924 12.751C14.3581 12.5784 14.4838 12.4378 14.7352 12.1565L15.0985 11.75" />
      <path d="M18.745 2.75V8.75M21.75 5.745L15.75 5.745" />
    </svg>
  );
};
