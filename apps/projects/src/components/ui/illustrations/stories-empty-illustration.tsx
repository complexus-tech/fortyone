import { cn } from "lib";

const SQUIRCLE_FRAME_PATH =
  "M62 34H218C230.8 34 234 37.2 234 50V140C234 152.8 230.8 156 218 156H62C49.2 156 46 152.8 46 140V50C46 37.2 49.2 34 62 34Z";

export const StoriesEmptyIllustration = ({
  className,
}: {
  className?: string;
}) => (
  <svg
    aria-hidden="true"
    className={cn("text-primary h-auto w-64", className)}
    fill="none"
    viewBox="0 0 280 180"
    xmlns="http://www.w3.org/2000/svg"
  >
    <circle cx="140" cy="90" fill="currentColor" opacity="0.04" r="86" />
    <path d={SQUIRCLE_FRAME_PATH} fill="currentColor" opacity="0.035" />
    <path
      d={SQUIRCLE_FRAME_PATH}
      stroke="currentColor"
      strokeOpacity="0.24"
      strokeWidth="1.5"
    />
    <path d="M46 61H234" stroke="currentColor" strokeOpacity="0.18" />
    <circle cx="61" cy="48" fill="currentColor" opacity="0.45" r="3" />
    <circle cx="72" cy="48" fill="currentColor" opacity="0.22" r="3" />
    <circle cx="83" cy="48" fill="currentColor" opacity="0.12" r="3" />
    <path d="M109 61V156" stroke="currentColor" strokeOpacity="0.12" />
    <path d="M171 61V156" stroke="currentColor" strokeOpacity="0.12" />
    <path
      d="M60 76H79"
      stroke="currentColor"
      strokeLinecap="round"
      strokeOpacity="0.5"
      strokeWidth="3"
    />
    <path
      d="M122 76H141"
      stroke="currentColor"
      strokeLinecap="round"
      strokeOpacity="0.34"
      strokeWidth="3"
    />
    <path
      d="M184 76H203"
      stroke="currentColor"
      strokeLinecap="round"
      strokeOpacity="0.2"
      strokeWidth="3"
    />
    <rect
      height="48"
      rx="8"
      stroke="currentColor"
      strokeDasharray="4 5"
      strokeOpacity="0.2"
      width="42"
      x="60"
      y="91"
    />
    <rect
      height="48"
      rx="8"
      stroke="currentColor"
      strokeDasharray="4 5"
      strokeOpacity="0.28"
      width="42"
      x="119"
      y="91"
    />
    <rect
      height="48"
      rx="8"
      stroke="currentColor"
      strokeDasharray="4 5"
      strokeOpacity="0.14"
      width="42"
      x="178"
      y="91"
    />
    <circle cx="140" cy="115" fill="currentColor" opacity="0.12" r="13" />
    <path
      d="M140 108V122M133 115H147"
      stroke="currentColor"
      strokeLinecap="round"
      strokeWidth="1.75"
    />
    <path
      d="M236 25L238.2 30.8L244 33L238.2 35.2L236 41L233.8 35.2L228 33L233.8 30.8L236 25Z"
      fill="currentColor"
      opacity="0.55"
    />
    <path
      d="M37 119L38.5 123L42.5 124.5L38.5 126L37 130L35.5 126L31.5 124.5L35.5 123L37 119Z"
      fill="currentColor"
      opacity="0.28"
    />
    <path
      d="M219 144C231 142 238 135 242 124"
      stroke="currentColor"
      strokeLinecap="round"
      strokeOpacity="0.32"
      strokeWidth="1.5"
    />
    <path
      d="M242 124L237 128M242 124L243 130"
      stroke="currentColor"
      strokeLinecap="round"
      strokeOpacity="0.32"
      strokeWidth="1.5"
    />
  </svg>
);
