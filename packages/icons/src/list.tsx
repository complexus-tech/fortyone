import type { Icon } from "./types";

export const ListIcon = (props: Icon) => {
  const { strokeWidth = 2, ...rest } = props;
  return (
    <svg
      {...rest}
      fill="currentColor"
      fillOpacity={0.1}
      strokeWidth={strokeWidth}
      height="24"
      viewBox="0 0 24 24"
      width="24"
      xmlns="http://www.w3.org/2000/svg"
    >
      <path
        d="M4 5L14 5"
        stroke="currentColor"
        strokeLinecap="round"
        strokeLinejoin="round"
      />
      <path
        d="M4 12L20 12"
        stroke="currentColor"
        strokeLinecap="round"
        strokeLinejoin="round"
      />
      <path
        d="M4 19L20 19"
        stroke="currentColor"
        strokeLinecap="round"
        strokeLinejoin="round"
      />
    </svg>
  );
};
