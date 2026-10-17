import { getMessageDate } from "./format";

const DATE_FORMAT = new Intl.DateTimeFormat("en", {
  weekday: "long",
  month: "long",
  day: "numeric",
  timeZone: "UTC",
});
const formatDate = (date: string) => {
  const now = new Date();
  const yesterday = new Date(now);
  yesterday.setDate(now.getDate() - 1);
  if (date === getMessageDate(now.toISOString())) return "Today";
  if (date === getMessageDate(yesterday.toISOString())) return "Yesterday";
  return DATE_FORMAT.format(new Date(`${date}T12:00:00Z`));
};

export const MessageDateDivider = ({ date }: { date: string }) => {
  const label = formatDate(date);

  return (
    <div className="relative flex items-center justify-center py-2">
      <time
        className="border-border bg-surface relative rounded-full border px-4 py-1 font-semibold"
        dateTime={date}
      >
        {label}
      </time>
    </div>
  );
};
