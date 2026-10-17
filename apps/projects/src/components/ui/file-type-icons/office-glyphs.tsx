"use client";

/**
 * Office file icons from vscode-icons. Branded artwork belongs to Microsoft.
 * Geometry and colors are unchanged; SVG IDs are unique per rendered instance.
 * Source files, pinned revision, and licensing notice are in ./NOTICE.md.
 */
import { useId } from "react";

export const WordFileGlyph = () => {
  const gradientId = useId();
  return (
    <svg
      aria-hidden
      className="size-4/5"
      fill="#000"
      focusable="false"
      viewBox="0 0 32 32"
    >
      <defs>
        <linearGradient
          gradientTransform="translate(0 1720)"
          gradientUnits="userSpaceOnUse"
          id={gradientId}
          x1="4.494"
          x2="13.832"
          y1="-1712.086"
          y2="-1695.914"
        >
          <stop offset="0" stopColor="#2368c4" />
          <stop offset=".5" stopColor="#1a5dbe" />
          <stop offset="1" stopColor="#1146ac" />
        </linearGradient>
      </defs>
      <path
        d="M28.806 3H9.705a1.19 1.19 0 0 0-1.193 1.191V9.5l11.069 3.25L30 9.5V4.191A1.19 1.19 0 0 0 28.806 3"
        fill="#41a5ee"
      />
      <path d="M30 9.5H8.512V16l11.069 1.95L30 16Z" fill="#2b7cd3" />
      <path d="M8.512 16v6.5l10.418 1.3L30 22.5V16Z" fill="#185abd" />
      <path
        d="M9.705 29h19.1A1.19 1.19 0 0 0 30 27.809V22.5H8.512v5.309A1.19 1.19 0 0 0 9.705 29"
        fill="#103f91"
      />
      <path
        d="M16.434 8.2H8.512v16.25h7.922a1.2 1.2 0 0 0 1.194-1.191V9.391A1.2 1.2 0 0 0 16.434 8.2"
        opacity=".10000000149011612"
        style={{ isolation: "isolate" }}
      />
      <path
        d="M15.783 8.85H8.512V25.1h7.271a1.2 1.2 0 0 0 1.194-1.191V10.041a1.2 1.2 0 0 0-1.194-1.191"
        opacity=".20000000298023224"
        style={{ isolation: "isolate" }}
      />
      <path
        d="M15.783 8.85H8.512V23.8h7.271a1.2 1.2 0 0 0 1.194-1.191V10.041a1.2 1.2 0 0 0-1.194-1.191"
        opacity=".20000000298023224"
        style={{ isolation: "isolate" }}
      />
      <path
        d="M15.132 8.85h-6.62V23.8h6.62a1.2 1.2 0 0 0 1.194-1.191V10.041a1.2 1.2 0 0 0-1.194-1.191"
        opacity=".20000000298023224"
        style={{ isolation: "isolate" }}
      />
      <path
        d="M3.194 8.85h11.938a1.193 1.193 0 0 1 1.194 1.191v11.918a1.193 1.193 0 0 1-1.194 1.191H3.194A1.19 1.19 0 0 1 2 21.959V10.041A1.19 1.19 0 0 1 3.194 8.85"
        fill={`url(#${gradientId})`}
      />
      <path
        d="M6.9 17.988q.035.276.046.481h.028q.015-.195.065-.47c.05-.275.062-.338.089-.465l1.255-5.407h1.624l1.3 5.326a8 8 0 0 1 .162 1h.022a8 8 0 0 1 .135-.975l1.039-5.358h1.477l-1.824 7.748h-1.727l-1.237-5.126q-.054-.222-.122-.578t-.084-.52h-.021q-.021.189-.084.561t-.1.552L7.78 19.871H6.024L4.19 12.127h1.5l1.131 5.418a5 5 0 0 1 .079.443"
        fill="#fff"
      />
    </svg>
  );
};

export const ExcelFileGlyph = () => {
  const gradientId = useId();
  return (
    <svg
      aria-hidden
      className="size-4/5"
      fill="#000"
      focusable="false"
      viewBox="0 0 32 32"
    >
      <defs>
        <linearGradient
          gradientTransform="translate(0 2100)"
          gradientUnits="userSpaceOnUse"
          id={gradientId}
          x1="4.494"
          x2="13.832"
          y1="-2092.086"
          y2="-2075.914"
        >
          <stop offset="0" stopColor="#18884f" />
          <stop offset=".5" stopColor="#117e43" />
          <stop offset="1" stopColor="#0b6631" />
        </linearGradient>
      </defs>
      <path
        d="M19.581 15.35 8.512 13.4v14.409A1.19 1.19 0 0 0 9.705 29h19.1A1.19 1.19 0 0 0 30 27.809V22.5Z"
        fill="#185c37"
      />
      <path
        d="M19.581 3H9.705a1.19 1.19 0 0 0-1.193 1.191V9.5L19.581 16l5.861 1.95L30 16V9.5Z"
        fill="#21a366"
      />
      <path d="M8.512 9.5h11.069V16H8.512Z" fill="#107c41" />
      <path
        d="M16.434 8.2H8.512v16.25h7.922a1.2 1.2 0 0 0 1.194-1.191V9.391A1.2 1.2 0 0 0 16.434 8.2"
        opacity=".10000000149011612"
        style={{ isolation: "isolate" }}
      />
      <path
        d="M15.783 8.85H8.512V25.1h7.271a1.2 1.2 0 0 0 1.194-1.191V10.041a1.2 1.2 0 0 0-1.194-1.191"
        opacity=".20000000298023224"
        style={{ isolation: "isolate" }}
      />
      <path
        d="M15.783 8.85H8.512V23.8h7.271a1.2 1.2 0 0 0 1.194-1.191V10.041a1.2 1.2 0 0 0-1.194-1.191"
        opacity=".20000000298023224"
        style={{ isolation: "isolate" }}
      />
      <path
        d="M15.132 8.85h-6.62V23.8h6.62a1.2 1.2 0 0 0 1.194-1.191V10.041a1.2 1.2 0 0 0-1.194-1.191"
        opacity=".20000000298023224"
        style={{ isolation: "isolate" }}
      />
      <path
        d="M3.194 8.85h11.938a1.193 1.193 0 0 1 1.194 1.191v11.918a1.193 1.193 0 0 1-1.194 1.191H3.194A1.19 1.19 0 0 1 2 21.959V10.041A1.19 1.19 0 0 1 3.194 8.85"
        fill={`url(#${gradientId})`}
      />
      <path
        d="m5.7 19.873 2.511-3.884-2.3-3.862h1.847L9.013 14.6c.116.234.2.408.238.524h.017q.123-.281.26-.546l1.342-2.447h1.7l-2.359 3.84 2.419 3.905h-1.809l-1.45-2.711A2.4 2.4 0 0 1 9.2 16.8h-.024a1.7 1.7 0 0 1-.168.351l-1.493 2.722Z"
        fill="#fff"
      />
      <path
        d="M28.806 3h-9.225v6.5H30V4.191A1.19 1.19 0 0 0 28.806 3"
        fill="#33c481"
      />
      <path d="M19.581 16H30v6.5H19.581Z" fill="#107c41" />
    </svg>
  );
};

export const PowerPointFileGlyph = () => {
  const gradientId = useId();
  return (
    <svg
      aria-hidden
      className="size-4/5"
      fill="#000"
      focusable="false"
      viewBox="0 0 32 32"
    >
      <defs>
        <linearGradient
          gradientTransform="translate(0 1756)"
          gradientUnits="userSpaceOnUse"
          id={gradientId}
          x1="4.494"
          x2="13.832"
          y1="-1748.086"
          y2="-1731.914"
        >
          <stop offset="0" stopColor="#ca4c28" />
          <stop offset=".5" stopColor="#c5401e" />
          <stop offset="1" stopColor="#b62f14" />
        </linearGradient>
      </defs>
      <path
        d="M18.93 17.3 16.977 3h-.146A12.9 12.9 0 0 0 3.953 15.854V16Z"
        fill="#ed6c47"
      />
      <path
        d="M17.123 3h-.146v13l6.511 2.6L30 16v-.146A12.9 12.9 0 0 0 17.123 3"
        fill="#ff8f6b"
      />
      <path
        d="M30 16v.143A12.905 12.905 0 0 1 17.12 29h-.287a12.907 12.907 0 0 1-12.88-12.857V16Z"
        fill="#d35230"
      />
      <path
        d="M17.628 9.389V23.26a1.2 1.2 0 0 1-.742 1.1 1.2 1.2 0 0 1-.45.091H7.027a10 10 0 0 1-.521-.65 12.74 12.74 0 0 1-2.553-7.657v-.286A12.7 12.7 0 0 1 6.05 8.85a9 9 0 0 1 .456-.65h9.93a1.2 1.2 0 0 1 1.192 1.189"
        opacity=".10000000149011612"
        style={{ isolation: "isolate" }}
      />
      <path
        d="M16.977 10.04v13.871a1.2 1.2 0 0 1-.091.448 1.2 1.2 0 0 1-1.1.741H7.62q-.309-.314-.593-.65a10 10 0 0 1-.521-.65 12.74 12.74 0 0 1-2.553-7.657v-.286A12.7 12.7 0 0 1 6.05 8.85h9.735a1.2 1.2 0 0 1 1.192 1.19"
        opacity=".20000000298023224"
        style={{ isolation: "isolate" }}
      />
      <path
        d="M16.977 10.04v12.571a1.2 1.2 0 0 1-1.192 1.189H6.506a12.74 12.74 0 0 1-2.553-7.657v-.286A12.7 12.7 0 0 1 6.05 8.85h9.735a1.2 1.2 0 0 1 1.192 1.19"
        opacity=".20000000298023224"
        style={{ isolation: "isolate" }}
      />
      <path
        d="M16.326 10.04v12.571a1.2 1.2 0 0 1-1.192 1.189H6.506a12.74 12.74 0 0 1-2.553-7.657v-.286A12.7 12.7 0 0 1 6.05 8.85h9.084a1.2 1.2 0 0 1 1.192 1.19"
        opacity=".20000000298023224"
        style={{ isolation: "isolate" }}
      />
      <path
        d="M3.194 8.85h11.938a1.193 1.193 0 0 1 1.194 1.191v11.918a1.193 1.193 0 0 1-1.194 1.191H3.194A1.19 1.19 0 0 1 2 21.959V10.041A1.19 1.19 0 0 1 3.194 8.85"
        fill={`url(#${gradientId})`}
      />
      <path
        d="M9.293 12.028a3.3 3.3 0 0 1 2.174.636 2.27 2.27 0 0 1 .756 1.841 2.56 2.56 0 0 1-.373 1.376 2.5 2.5 0 0 1-1.059.935 3.6 3.6 0 0 1-1.591.334H7.687v2.8H6.141v-7.922ZM7.686 15.94h1.331a1.74 1.74 0 0 0 1.177-.351 1.3 1.3 0 0 0 .4-1.025q0-1.309-1.525-1.31H7.686z"
        fill="#fff"
      />
    </svg>
  );
};
