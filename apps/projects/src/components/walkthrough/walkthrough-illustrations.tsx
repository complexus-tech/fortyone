type WalkthroughStartChoice = "calendar" | "maya" | "objective" | "task";

export const WalkthroughStartChoiceIllustration = ({
  className,
  choice,
}: {
  choice: WalkthroughStartChoice;
  className?: string;
}) => (
  <svg
    aria-hidden="true"
    className={className}
    fill="none"
    viewBox="0 0 120 120"
    xmlns="http://www.w3.org/2000/svg"
  >
    <path
      d="M39 21C30 21 23 28 23 37C22 48 24 58 22.5 70C21.5 84 29 98 42 98C54 98.6 65 96.8 78 98.5C90 100 98 91 98 80C99 67 96.5 56 98.5 43C100 30 92 21 80 22C66 20.5 52 22 39 21Z"
      fill="none"
      stroke="currentColor"
      strokeLinecap="round"
      strokeLinejoin="round"
      strokeOpacity="0.52"
      strokeWidth="2"
    />

    {choice === "task" ? (
      <>
        <path
          d="M43 47H77M43 58H68"
          stroke="currentColor"
          strokeLinecap="round"
          strokeOpacity="0.72"
          strokeWidth="3"
        />
        <rect
          height="14"
          rx="4"
          stroke="currentColor"
          strokeOpacity="0.62"
          strokeWidth="2"
          width="14"
          x="42"
          y="72"
        />
        <path
          d="M62 79H78"
          stroke="currentColor"
          strokeLinecap="round"
          strokeOpacity="0.45"
          strokeWidth="3"
        />
        <path
          d="M46 79L49 82L54 76"
          stroke="currentColor"
          strokeLinecap="round"
          strokeLinejoin="round"
          strokeOpacity="0.72"
          strokeWidth="2"
        />
      </>
    ) : null}

    {choice === "objective" ? (
      <>
        <circle
          cx="60"
          cy="62"
          r="22"
          stroke="currentColor"
          strokeOpacity="0.3"
          strokeWidth="2"
        />
        <circle
          cx="60"
          cy="62"
          r="12"
          stroke="currentColor"
          strokeOpacity="0.54"
          strokeWidth="2"
        />
        <circle cx="60" cy="62" fill="currentColor" opacity="0.8" r="4" />
        <path
          d="M74 47L83 38M77 38H83V44"
          stroke="currentColor"
          strokeLinecap="round"
          strokeLinejoin="round"
          strokeOpacity="0.78"
          strokeWidth="2.5"
        />
        <path
          d="M39 85H81"
          stroke="currentColor"
          strokeLinecap="round"
          strokeOpacity="0.24"
          strokeWidth="2"
        />
      </>
    ) : null}

    {choice === "calendar" ? (
      <>
        <path
          d="M20 46H100M43 20V32M77 20V32"
          stroke="currentColor"
          strokeOpacity="0.54"
          strokeWidth="2"
        />
        <path
          d="M39 60H48M56 60H65M73 60H82M39 76H48M56 76H65"
          stroke="currentColor"
          strokeLinecap="round"
          strokeOpacity="0.38"
          strokeWidth="5"
        />
        <rect
          fill="currentColor"
          height="13"
          opacity="0.7"
          rx="3"
          width="13"
          x="72"
          y="69"
        />
      </>
    ) : null}

    {choice === "maya" ? (
      <>
        <path
          d="M38 46C38 40.5 42.5 36 48 36H72C77.5 36 82 40.5 82 46V65C82 70.5 77.5 75 72 75H58L48 84V75C42.5 75 38 70.5 38 65V46Z"
          stroke="currentColor"
          strokeLinecap="round"
          strokeLinejoin="round"
          strokeOpacity="0.68"
          strokeWidth="2.5"
        />
        <path
          d="M49 51H71M49 61H64"
          stroke="currentColor"
          strokeLinecap="round"
          strokeOpacity="0.44"
          strokeWidth="3"
        />
        <path
          d="M84 45L86 50L91 52L86 54L84 59L82 54L77 52L82 50L84 45Z"
          fill="currentColor"
          opacity="0.8"
        />
      </>
    ) : null}
  </svg>
);
