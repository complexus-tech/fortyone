import type { Icon } from "@/types/icon";

export const MinimizeIcon = (props: Icon) => {
  return (
    <svg
      {...props}
      fill="none"
      height="24"
      viewBox="0 0 24 24"
      width="24"
      xmlns="http://www.w3.org/2000/svg"
    >
      <path
        d="M6.5022 13.2635C7.34661 13.2515 10.1431 12.6706 10.736 13.2635C11.3289 13.8564 10.748 16.6529 10.736 17.4973M13.2684 6.49733C13.2564 7.34173 12.6755 10.1382 13.2684 10.7311C13.8613 11.324 16.6578 10.7431 17.5022 10.7311M20.999 2.99902L13.6101 10.3812M10.369 13.6237L2.99987 21.001"
        stroke="currentColor"
        strokeLinecap="round"
        strokeLinejoin="round"
        strokeWidth="2"
      />
    </svg>
  );
};
