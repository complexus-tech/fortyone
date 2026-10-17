import { AtIcon, CalendarIcon, ChatIcon, CommentIcon } from "icons";
import { Dot } from "@/components/ui/dot";
import { PriorityIcon } from "@/components/ui/priority-icon";
import type { AppNotification } from "./types";
import { getNotificationIndicator } from "./utils/notification-indicator";

export const NotificationIndicator = (
  notification: Pick<AppNotification, "type" | "message">,
) => {
  const indicator = getNotificationIndicator(notification);
  if (!indicator) return null;

  switch (indicator.kind) {
    case "calendar":
      return <CalendarIcon className="shrink-0" />;
    case "comment":
      return <CommentIcon className="shrink-0" />;
    case "mention":
      return <AtIcon className="shrink-0" />;
    case "chat":
      return <ChatIcon className="shrink-0" />;
    case "status":
      return <Dot className="text-text-muted size-3" />;
    case "priority":
      return <PriorityIcon priority={indicator.priority} />;
  }
};
