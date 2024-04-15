import type { Icon } from "./types";

export const MenuIcon = (props: Icon) => {
  const { strokeWidth = 2, ...rest } = props;
  return (
    <svg
      {...rest}
      fill="none"
      height="24"
      strokeWidth={strokeWidth}
      viewBox="0 0 24 24"
      width="24"
      xmlns="http://www.w3.org/2000/svg"
    >
      <path
        d="M4 8.5L20 8.5"
        stroke="currentColor"
        strokeLinecap="round"
        strokeLinejoin="round"
      />
      <path
        d="M4 15.5L20 15.5"
        stroke="currentColor"
        strokeLinecap="round"
        strokeLinejoin="round"
      />
    </svg>
  );
};
