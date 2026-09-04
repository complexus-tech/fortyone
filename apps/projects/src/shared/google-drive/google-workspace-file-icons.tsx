import type { ComponentPropsWithoutRef } from "react";
import { useId } from "react";
import { cn } from "lib";

type GoogleWorkspaceFileIconProps = ComponentPropsWithoutRef<"svg">;

export const GoogleDocsIcon = ({
  className,
  ...props
}: GoogleWorkspaceFileIconProps) => {
  const instanceId = useId().replaceAll(":", "");
  const filterId = `${instanceId}-google-docs-filter`;
  const gradientId = `${instanceId}-google-docs-gradient`;
  const maskId = `${instanceId}-google-docs-mask`;

  return (
    <svg
      aria-hidden="true"
      className={cn("size-5 shrink-0", className)}
      fill="none"
      viewBox="0 0 192 192"
      xmlns="http://www.w3.org/2000/svg"
      {...props}
    >
      <mask
        height="176"
        id={maskId}
        maskUnits="userSpaceOnUse"
        style={{ maskType: "alpha" }}
        width="128"
        x="32"
        y="8"
      >
        <path
          d="M130.334 184H61.6c-8.9435 0-13.4152 0-16.9625-1.404a20 20 0 0 1-11.233-11.234C32 167.815 32 163.343 32 154.4V37.6c0-8.9435 0-13.4152 1.4045-16.9625a20 20 0 0 1 11.233-11.233C48.1848 8 52.6565 8 61.6 8H100l54.793 54.7933v.0001c1.661 1.6609 2.492 2.4914 3.13 3.4305a11.99 11.99 0 0 1 1.862 4.5011c.212 1.1154.212 2.3014.21 4.6735-.035 48.9185-.057 49.4005-.058 78.9675 0 8.966 0 13.45-1.405 16.997a20 20 0 0 1-11.233 11.233C143.752 184 139.279 184 130.334 184"
          fill="#3186FF"
        />
      </mask>
      <g mask={`url(#${maskId})`}>
        <path d="M159.94 184H31.9999V8h68l59.9991 60z" fill="#3186FF" />
        <g filter={`url(#${filterId})`}>
          <path d="M43 192h106V20H43Z" fill={`url(#${gradientId})`} />
        </g>
      </g>
      <path
        d="M154.995 62.995A15 15 0 0 0 146 60h-33.2c-7.069 0-12.8-5.731-12.8-12.8V8Z"
        fill="#76BBFF"
      />
      <rect fill="white" height="12" rx="6" width="64" x="64.001" y="114" />
      <rect fill="white" height="12" rx="6" width="48" x="64.001" y="143" />
      <defs>
        <filter
          colorInterpolationFilters="sRGB"
          filterUnits="userSpaceOnUse"
          height="196"
          id={filterId}
          width="130"
          x="31"
          y="8"
        >
          <feFlood floodOpacity="0" result="BackgroundImageFix" />
          <feBlend
            in="SourceGraphic"
            in2="BackgroundImageFix"
            mode="normal"
            result="shape"
          />
          <feGaussianBlur
            result="effect1_foregroundBlur_37242_8762"
            stdDeviation="6"
          />
        </filter>
        <linearGradient
          gradientUnits="userSpaceOnUse"
          id={gradientId}
          x1="96"
          x2="54.6124"
          y1="59.2839"
          y2="171.338"
        >
          <stop offset=".33" stopColor="#3186FF" />
          <stop offset="1" stopColor="#A9A8FF" />
        </linearGradient>
      </defs>
    </svg>
  );
};

export const GoogleSheetsIcon = ({
  className,
  ...props
}: GoogleWorkspaceFileIconProps) => {
  const instanceId = useId().replaceAll(":", "");
  const filterId = `${instanceId}-google-sheets-filter`;
  const gradientId = `${instanceId}-google-sheets-gradient`;
  const maskId = `${instanceId}-google-sheets-mask`;

  return (
    <svg
      aria-hidden="true"
      className={cn("size-5 shrink-0", className)}
      fill="none"
      viewBox="0 0 192 192"
      xmlns="http://www.w3.org/2000/svg"
      {...props}
    >
      <path
        d="M183.96 117.4c0 8.94 0 13.41-1.4 16.96a20 20 0 0 1-11.23 11.24c-3.55 1.4-8.02 1.4-16.97 1.4h-60.8c-8.94 0-13.41 0-16.96-1.4a20 20 0 0 1-11.23-11.24c-1.4-3.55-1.4-8.02-1.4-16.96V74.6c0-8.94 0-13.42 1.4-16.96A20 20 0 0 1 76.6 46.4C80.15 45 84.62 45 93.56 45h60.8c8.95 0 13.42 0 16.97 1.4a20 20 0 0 1 11.23 11.24c1.4 3.54 1.4 8.02 1.4 16.96z"
        fill="#009954"
      />
      <mask
        height="128"
        id={maskId}
        maskUnits="userSpaceOnUse"
        style={{ maskType: "alpha" }}
        width="161"
        x="7"
        y="32"
      >
        <path
          d="M167.96 130.4c0 8.94 0 13.41-1.4 16.96a20 20 0 0 1-11.23 11.24c-3.55 1.4-8.02 1.4-16.97 1.4H37.56c-8.94 0-13.41 0-16.96-1.4a20 20 0 0 1-11.23-11.24c-1.4-3.55-1.4-8.02-1.4-16.96V61.6c0-8.94 0-13.42 1.4-16.96A20 20 0 0 1 20.6 33.4C24.15 32 28.62 32 37.56 32h100.8c8.95 0 13.42 0 16.97 1.4a20 20 0 0 1 11.23 11.24c1.4 3.54 1.4 8.02 1.4 16.96z"
          fill="#0EBC5F"
        />
      </mask>
      <g mask={`url(#${maskId})`}>
        <path d="M167.96 160h-160V32h160z" fill="#0EBC5F" />
        <g filter={`url(#${filterId})`}>
          <path
            d="M183.96 65a20 20 0 0 0-20-20h-104a20 20 0 0 0-20 20v62a20 20 0 0 0 20 20h104a20 20 0 0 0 20-20z"
            fill={`url(#${gradientId})`}
          />
        </g>
      </g>
      <path
        d="M47.96 122a6 6 0 0 1-6-6V77h-14a6 6 0 0 1 0-12h14V52a6 6 0 0 1 12 0v13h58a6 6 0 1 1 0 12h-58v39a6 6 0 0 1-6 6"
        fill="#fff"
      />
      <defs>
        <linearGradient
          gradientUnits="userSpaceOnUse"
          id={gradientId}
          x1="61.73"
          x2="163.21"
          y1="88.31"
          y2="88.31"
        >
          <stop stopColor="#0EBC5F" />
          <stop offset=".95" stopColor="#78C9FF" />
        </linearGradient>
        <filter
          colorInterpolationFilters="sRGB"
          filterUnits="userSpaceOnUse"
          height="126"
          id={filterId}
          width="168"
          x="27.96"
          y="33"
        >
          <feFlood floodOpacity="0" result="BackgroundImageFix" />
          <feBlend in="SourceGraphic" in2="BackgroundImageFix" result="shape" />
          <feGaussianBlur
            result="effect1_foregroundBlur_2015_872"
            stdDeviation="6"
          />
        </filter>
      </defs>
    </svg>
  );
};
