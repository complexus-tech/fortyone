import type { Icon } from "./types";

export const ArrowLeftIcon = (props: Icon) => {
  return (
    <svg
      {...props}
      fill="currentColor"
      fillOpacity={0.1}
      height="24"
      viewBox="0 0 24 24"
      width="24"
      xmlns="http://www.w3.org/2000/svg"
    >
      <path
        d="M15 6C15 6 9.00001 10.4189 9 12C8.99999 13.5812 15 18 15 18"
        stroke="currentColor"
        strokeLinecap="round"
        strokeLinejoin="round"
        strokeWidth="3"
      />
    </svg>
  );
};
