import type { Icon } from "./types";

export const ArrowUpIcon = (props: Icon) => {
  const { strokeWidth = 3, ...rest } = props;
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
        d="M18 15C18 15 13.5811 9.00001 12 9C10.4188 8.99999 6 15 6 15"
        stroke="currentColor"
        strokeLinecap="round"
        strokeLinejoin="round"
        strokeWidth="3"
      />
    </svg>
  );
};
