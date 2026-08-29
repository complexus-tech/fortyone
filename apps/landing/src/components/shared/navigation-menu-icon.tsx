import type { ReactNode } from "react";
import { cn } from "lib";

export type NavigationIconName =
  | "ai-planning"
  | "api-reference"
  | "blog"
  | "calendar"
  | "customer-feedback"
  | "customer-support"
  | "developers"
  | "documents"
  | "goals"
  | "government"
  | "integrations"
  | "marketing"
  | "operations"
  | "pitch"
  | "product"
  | "roadmaps"
  | "strategy-map"
  | "tasks";

export type NavigationIconTone =
  | "aqua"
  | "blue"
  | "lime"
  | "lilac"
  | "orange"
  | "rose";

const toneClasses: Record<NavigationIconTone, string> = {
  aqua: "bg-[#d8f2ee] text-[#174a47] dark:bg-[#183b3a] dark:text-[#a9e1dc]",
  blue: "bg-[#dcecff] text-[#173e69] dark:bg-[#172f4b] dark:text-[#a9cef7]",
  lime: "bg-[#eef7ce] text-[#3c4e17] dark:bg-[#303b1b] dark:text-[#d5e99a]",
  lilac: "bg-[#e9dfff] text-[#49316f] dark:bg-[#33264d] dark:text-[#cfbafa]",
  orange: "bg-[#ffe3c3] text-[#684012] dark:bg-[#49301a] dark:text-[#f4c68e]",
  rose: "bg-[#f8e0e8] text-[#672a41] dark:bg-[#482535] dark:text-[#efb6c9]",
};

const glyphs: Record<NavigationIconName, ReactNode> = {
  "ai-planning": (
    <>
      <path d="m10.9 3 .5 1.7a4 4 0 0 0 2.7 2.7l1.7.5-1.7.5a4 4 0 0 0-2.7 2.7l-.5 1.7-.5-1.7a4 4 0 0 0-2.7-2.7L6 7.9l1.7-.5a4 4 0 0 0 2.7-2.7L10.9 3Z" />
      <path d="m5.3 13.1.3.9a2.2 2.2 0 0 0 1.5 1.5l.9.3-.9.3a2.2 2.2 0 0 0-1.5 1.5l-.3.9-.3-.9a2.2 2.2 0 0 0-1.5-1.5l-.9-.3.9-.3A2.2 2.2 0 0 0 5 14l.3-.9Z" />
      <path d="m14.5 14.1.2.6a1.6 1.6 0 0 0 1.1 1.1l.6.2-.6.2a1.6 1.6 0 0 0-1.1 1.1l-.2.6-.2-.6a1.6 1.6 0 0 0-1.1-1.1l-.6-.2.6-.2a1.6 1.6 0 0 0 1.1-1.1l.2-.6Z" />
    </>
  ),
  "api-reference": (
    <>
      <path d="m7.2 5.5-4 4.5 4 4.5M12.8 5.5l4 4.5-4 4.5" />
      <path d="m11.3 3.5-2.6 13" />
    </>
  ),
  blog: (
    <>
      <path d="M4 4.5h12v11H4z" />
      <path d="M7 7.5h6M7 10.2h6M7 12.9h3.5" />
    </>
  ),
  calendar: (
    <>
      <rect height="13" rx="2" width="14" x="3" y="4" />
      <path d="M6.5 2.8v2.6M13.5 2.8v2.6M3 8h14" />
      <path d="m7 12 1.5 1.5L13 10" />
    </>
  ),
  "customer-feedback": (
    <g transform="scale(.833333)">
      <path d="M5 4h14a2 2 0 0 1 2 2v9a2 2 0 0 1-2 2h-8l-6 4v-4a2 2 0 0 1-2-2V6a2 2 0 0 1 2-2Z" />
      <path d="M8 9h8" />
      <path d="M8 13h5" />
    </g>
  ),
  "customer-support": (
    <>
      <path d="M4 11v-1a6 6 0 0 1 12 0v1" />
      <path d="M4 10.5H3.5A1.5 1.5 0 0 0 2 12v2a1.5 1.5 0 0 0 1.5 1.5H5V10.5H4ZM16 10.5h.5A1.5 1.5 0 0 1 18 12v2a1.5 1.5 0 0 1-1.5 1.5H15v-5h1Z" />
      <path d="M15 15.5c-.7 1.1-1.8 1.7-3.4 1.7H10" />
    </>
  ),
  developers: <path d="m7.5 6-4 4 4 4M12.5 6l4 4-4 4M11 3.5 9 16.5" />,
  documents: (
    <>
      <path d="M6 3.5h6l3 3v10H6z" />
      <path d="M12 3.5v3h3M8.5 10h4M8.5 12.7h4" />
      <path d="M4 6.5v10h8" />
    </>
  ),
  goals: (
    <>
      <circle cx="9" cy="11" r="6" />
      <circle cx="9" cy="11" r="2.5" />
      <path d="m10.7 9.3 5.8-5.8M13.2 3.5h3.3v3.3" />
    </>
  ),
  government: (
    <path d="m10 2.8 7 3.7H3l7-3.7ZM4.5 8.5v6M8.2 8.5v6M11.8 8.5v6M15.5 8.5v6M3 17h14" />
  ),
  integrations: (
    <>
      <path d="M7.2 4.2v3.1H4.1M12.8 4.2v3.1h3.1M7.2 15.8v-3.1H4.1M12.8 15.8v-3.1h3.1" />
      <rect height="6" rx="1.5" width="6" x="7" y="7" />
    </>
  ),
  marketing: (
    <>
      <path d="M3.5 9v3.5h3l7 3.4V5.6l-7 3.4h-3Z" />
      <path d="m6.5 12.5 1.2 4h2.6M15.5 8.2c1 .9 1 2.7 0 3.6" />
    </>
  ),
  operations: (
    <>
      <path d="M3 5h14M3 10h14M3 15h14" />
      <circle cx="7" cy="5" r="1.5" />
      <circle cx="13" cy="10" r="1.5" />
      <circle cx="8.5" cy="15" r="1.5" />
    </>
  ),
  pitch: (
    <>
      <rect height="10" rx="1.5" width="14" x="3" y="3.5" />
      <path d="M10 13.5V17M7 17h6M6.5 10V8.2M10 10V6.5M13.5 10V7.4" />
    </>
  ),
  product: (
    <>
      <path d="m10 2.8 6.5 3.5v7.4L10 17.2l-6.5-3.5V6.3L10 2.8Z" />
      <path d="m3.5 6.3 6.5 3.5 6.5-3.5M10 9.8v7.4" />
    </>
  ),
  roadmaps: (
    <>
      <circle cx="4" cy="15.5" r="1.5" />
      <circle cx="10" cy="10" r="1.5" />
      <circle cx="16" cy="4.5" r="1.5" />
      <path d="M5.3 14.2 8.7 11.3M11.3 8.7l3.4-2.9" />
      <path d="M16 6v7h-3" />
    </>
  ),
  "strategy-map": (
    <>
      <circle cx="10" cy="4" r="2" />
      <circle cx="4.5" cy="15.5" r="2" />
      <circle cx="15.5" cy="15.5" r="2" />
      <path d="M9 5.8 5.5 13.7M11 5.8l3.5 7.9M6.5 15.5h7" />
    </>
  ),
  tasks: (
    <>
      <rect height="3" rx=".8" width="3" x="3" y="4" />
      <rect height="3" rx=".8" width="3" x="3" y="9" />
      <rect height="3" rx=".8" width="3" x="3" y="14" />
      <path d="M8.5 5.5H17M8.5 10.5H17M8.5 15.5H14" />
    </>
  ),
};

export const NavigationMenuIcon = ({
  className,
  name,
  tone,
}: {
  className?: string;
  name: NavigationIconName;
  tone: NavigationIconTone;
}) => (
  <span
    aria-hidden="true"
    className={cn(
      "flex size-10 shrink-0 items-center justify-center rounded-xl transition-transform duration-200 group-hover:scale-[1.04] motion-reduce:transition-none",
      toneClasses[tone],
      className,
    )}
  >
    <svg
      className="size-5"
      fill="none"
      stroke="currentColor"
      strokeLinecap="round"
      strokeLinejoin="round"
      strokeWidth="1.55"
      viewBox="0 0 20 20"
    >
      {glyphs[name]}
    </svg>
  </span>
);
