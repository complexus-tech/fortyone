import type { Icon } from "./types";

export const ArrowDownIcon = (props: Icon) => {
  const { strokeWidth = 3, ...rest } = props;
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
        d="M18 9.00005C18 9.00005 13.5811 15 12 15C10.4188 15 6 9 6 9"
        stroke="currentColor"
        strokeLinecap="round"
        strokeLinejoin="round"
        strokeWidth="3"
      />
    </svg>
  );
};
