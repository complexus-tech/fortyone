import { InternetIcon } from "icons";

const BrowserIcon = ({ name }: { name: string }) => {
  const iconClass = "text-text-muted size-4 shrink-0";
  if (!["Chrome", "Safari", "Opera", "Arc", "Dia"].includes(name))
    return <InternetIcon aria-hidden className={iconClass} />;

  return (
    <svg
      aria-hidden
      className={iconClass}
      fill="none"
      stroke="currentColor"
      strokeLinecap="round"
      strokeLinejoin="round"
      strokeWidth="1.6"
      viewBox="0 0 24 24"
    >
      {name === "Chrome" ? (
        <>
          <circle cx="12" cy="12" r="9" />
          <circle cx="12" cy="12" r="3.5" />
          <path d="M12 8.5h8M15 13.8l-4 7M9 13.8l-4-7" />
        </>
      ) : null}
      {name === "Safari" ? (
        <>
          <circle cx="12" cy="12" r="9" />
          <path d="m16 8-2.5 5.5L8 16l2.5-5.5L16 8ZM12 3v1M12 20v1M3 12h1M20 12h1" />
        </>
      ) : null}
      {name === "Opera" ? (
        <>
          <ellipse cx="12" cy="12" rx="8" ry="9" />
          <ellipse cx="12" cy="12" rx="4.5" ry="8.5" />
        </>
      ) : null}
      {name === "Arc" ? <path d="m4 20 8-17 8 17M7 13a7 7 0 0 0 10 0" /> : null}
      {name === "Dia" ? (
        <>
          <path d="m12 3 8 9-8 9-8-9 8-9Z" />
          <path d="M12 3v18M4 12h16" />
        </>
      ) : null}
    </svg>
  );
};

export const SessionBrowser = ({
  name,
  current,
}: {
  name?: string | null;
  current: boolean;
}) => {
  const label = name || "Unknown browser";
  return (
    <span
      className="inline-flex items-center gap-2 whitespace-nowrap"
      title={current ? `${label} · This browser` : label}
    >
      <BrowserIcon name={label} />
      {label}
      {current ? (
        <>
          <span aria-hidden className="bg-primary size-1.5 rounded-full" />
          <span className="sr-only"> (This browser)</span>
        </>
      ) : null}
    </span>
  );
};
