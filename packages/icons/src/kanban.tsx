import type { Icon } from "./types";

export const KanbanIcon = (props: Icon) => {
  const { strokeWidth = 2, ...rest } = props;
  return (
    <svg
      xmlns="http://www.w3.org/2000/svg"
      {...rest}
      fill="none"
      height="24"
      viewBox="0 0 24 24"
      width="24"
      strokeWidth={strokeWidth}
      stroke="currentColor"
      strokeLinecap="round"
      strokeLinejoin="round"
    >
      <rect width="18" height="18" x="3" y="3" rx="2" />
      <path d="M9 3v18" />
      <path d="M15 3v18" />
    </svg>
  );
};
