import type { ReactNode } from "react";

const ActionIcon = ({ children }: { children: ReactNode }) => (
  <svg
    aria-hidden="true"
    className="size-10 shrink-0"
    fill="none"
    stroke="currentColor"
    strokeLinecap="round"
    strokeLinejoin="round"
    strokeWidth={1.6}
    viewBox="0 0 40 40"
  >
    {children}
  </svg>
);

export const TaskActionIcon = () => (
  <ActionIcon>
    <rect
      fill="currentColor"
      fillOpacity={0.05}
      height="30"
      rx="5"
      width="24"
      x="8"
      y="5"
    />
    <path d="M14 13h12M14 18h8" opacity={0.4} />
    <path d="m13 26 3 3 6-7M25 27h2" />
  </ActionIcon>
);

export const ImportActionIcon = () => (
  <ActionIcon>
    <rect
      fill="currentColor"
      fillOpacity={0.05}
      height="20"
      rx="3"
      width="18"
      x="11"
      y="4"
    />
    <path d="M16 10h8M16 14h5" opacity={0.4} />
    <path d="M6 25v7a3 3 0 0 0 3 3h22a3 3 0 0 0 3-3v-7" />
    <path d="M20 20v10m-4-4 4 4 4-4" strokeWidth={2} />
  </ActionIcon>
);

export const CalendarActionIcon = () => (
  <ActionIcon>
    <rect
      fill="currentColor"
      fillOpacity={0.05}
      height="27"
      rx="5"
      width="28"
      x="6"
      y="8"
    />
    <path d="M6 16h28" opacity={0.4} />
    <path d="M13 5v6M27 5v6" />
    <path
      d="m13 24 3-2v8m6-6a2.5 2.5 0 0 1 5 0c0 2-5 3-5 6h5"
      strokeWidth={2}
    />
  </ActionIcon>
);
