import type { Icon } from "./types";

export const TwitterIcon = (props: Icon) => {
  const { strokeWidth = 2, ...rest } = props;
  return (
    <svg
      {...rest}
      fill="currentColor"
      strokeWidth={strokeWidth}
      fillOpacity={0.1}
      height="24"
      viewBox="0 0 24 24"
      width="24"
      xmlns="http://www.w3.org/2000/svg"
    >
      <path
        d="M3 21L10.5484 13.4516M21 3L13.4516 10.5484M13.4516 10.5484L8 3H3L10.5484 13.4516M13.4516 10.5484L21 21H16L10.5484 13.4516"
        stroke="currentColor"
        strokeLinecap="round"
        strokeLinejoin="round"
      />
    </svg>
  );
};
