import type { ComponentPropsWithoutRef } from "react";
import { useId } from "react";

export const GoogleDriveIcon = ({
  className,
  ...props
}: ComponentPropsWithoutRef<"svg">) => {
  const instanceId = useId().replaceAll(":", "");
  const maskId = `${instanceId}-google-drive-mask`;
  const rightGradientId = `${instanceId}-google-drive-right-gradient`;
  const bottomGradientId = `${instanceId}-google-drive-bottom-gradient`;
  const leftGradientId = `${instanceId}-google-drive-left-gradient`;

  return (
    <svg
      aria-hidden="true"
      className={className}
      fill="none"
      viewBox="0 0 192 192"
      xmlns="http://www.w3.org/2000/svg"
      {...props}
    >
      <mask
        height="154"
        id={maskId}
        maskUnits="userSpaceOnUse"
        style={{ maskType: "alpha" }}
        width="168"
        x="12"
        y="18"
      >
        <path
          d="M63.09 37c14.626-25.333 51.193-25.334 65.819 0l45.033 78c14.626 25.334-3.657 57.001-32.91 57.001H50.967c-29.253 0-47.536-31.667-32.91-57.001z"
          fill="#b43333"
        />
      </mask>
      <g mask={`url(#${maskId})`}>
        <path
          d="M206.905 172.02h-91.888l-19.015-32.934 45.944-79.578z"
          fill={`url(#${rightGradientId})`}
        />
        <path
          d="M-14.919 172.006 50.04 59.494v.002L31.032 92.422h38.02L115 172.004l-129.918.001z"
          fill={`url(#${bottomGradientId})`}
        />
        <path
          d="M96.007-20.085 141.954 59.5l-19.011 32.928H31.048z"
          fill={`url(#${leftGradientId})`}
        />
      </g>
      <defs>
        <linearGradient
          gradientUnits="userSpaceOnUse"
          id={rightGradientId}
          x1="193.6"
          x2="103.09"
          y1="165.6"
          y2="111.21"
        >
          <stop offset=".09" stopColor="#ffe921" />
          <stop offset="1" stopColor="#fec700" />
        </linearGradient>
        <linearGradient
          gradientUnits="userSpaceOnUse"
          id={bottomGradientId}
          x1="114.4"
          x2="15.53"
          y1="181.61"
          y2="121.8"
        >
          <stop offset=".15" stopColor="#a9a8ff" />
          <stop offset=".33" stopColor="#6d97ff" />
          <stop offset=".48" stopColor="#3186ff" />
        </linearGradient>
        <linearGradient
          gradientUnits="userSpaceOnUse"
          id={leftGradientId}
          x1="128.88"
          x2="28.7"
          y1="37.88"
          y2="84.64"
        >
          <stop offset=".55" stopColor="#0ebc5f" />
          <stop offset=".85" stopColor="#78c9ff" />
        </linearGradient>
      </defs>
    </svg>
  );
};
