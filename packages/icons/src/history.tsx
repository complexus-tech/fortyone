import { cn } from "lib";
import type { Icon } from "./types";

export const HistoryIcon = (props: Icon) => {
  const { strokeWidth = 2.5, className } = props;
  return (
    <svg
      {...props}
      className={cn("h-5 w-auto text-gray dark:text-gray-300", className)}
      fill="none"
      height="24"
      viewBox="0 0 24 24"
      strokeWidth={strokeWidth}
      width="24"
      xmlns="http://www.w3.org/2000/svg"
    >
      <path
        d="M5.04798 8.60657L2.53784 8.45376C4.33712 3.70477 9.503 0.999914 14.5396 2.34474C19.904 3.77711 23.0904 9.26107 21.6565 14.5935C20.2227 19.926 14.7116 23.0876 9.3472 21.6553C5.36419 20.5917 2.58192 17.2946 2 13.4844"
        stroke="currentColor"
        strokeLinecap="round"
        strokeLinejoin="round"
      />
      <path
        d="M12 8V12L14 14"
        stroke="currentColor"
        strokeLinecap="round"
        strokeLinejoin="round"
      />
    </svg>
  );
};
