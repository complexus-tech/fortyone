import { securityDate, securityDateParts } from "./display";

export const SecurityTimestamp = ({ value }: { value: string }) => {
  const parts = securityDateParts(value);
  return (
    <time
      aria-label={securityDate(value)}
      className="whitespace-nowrap tabular-nums"
      dateTime={value}
      title={securityDate(value)}
    >
      {parts.date} · {parts.time.replace(" UTC", "")}
    </time>
  );
};
