import { cn } from "lib";

export const ChatUnreadBadge = ({
  count,
  label = "unread messages",
  className,
}: {
  count: number;
  label?: string;
  className?: string;
}) => {
  if (count <= 0) return null;

  return (
    <span
      aria-label={`${count.toLocaleString()} ${label}`}
      className={cn(
        "bg-primary/10 text-primary inline-flex min-w-5 shrink-0 items-center justify-center rounded-full px-1.5 py-0.5 text-xs leading-4 font-semibold tabular-nums",
        className,
      )}
      title={`${count.toLocaleString()} ${label}`}
    >
      <span aria-hidden="true">{count > 999 ? "999+" : count}</span>
    </span>
  );
};
