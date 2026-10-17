import { cn } from "lib";
import { Badge } from "ui";

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
    <Badge
      aria-label={`${count.toLocaleString()} ${label}`}
      className={cn("shrink-0 tabular-nums", className)}
      rounded="full"
      size="sm"
      title={`${count.toLocaleString()} ${label}`}
    >
      <span aria-hidden="true">{count > 999 ? "999+" : count}</span>
    </Badge>
  );
};
