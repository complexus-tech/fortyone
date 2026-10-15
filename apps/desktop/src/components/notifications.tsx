import { useAccountNative } from "../hooks/use-account-native";
import { useState } from "react";
import type { AppNotification, Story, WorkspaceData } from "../types";
import { errorMessage } from "../native";
import { formatTime } from "../model";
import { Empty, ErrorNotice } from "./primitives";

function notificationMessage(notification: AppNotification) {
  if (typeof notification.message === "string") return notification.message;
  return (
    notification.message?.template?.replace(
      /\{\{?([\w]+)\}?\}/g,
      (token, key: string) =>
        notification.message && typeof notification.message !== "string"
          ? notification.message.variables?.[key]?.value ?? token
          : token,
    ) ?? ""
  );
}
export function Notifications({
  data,
  onStory,
  acceptData,
}: {
  data: WorkspaceData;
  onStory: (story: Story) => void;
  acceptData: (data: WorkspaceData) => void;
}) {
  const client = useAccountNative();
  const [error, setError] = useState<string | null>(null);
  const [busy, setBusy] = useState<string | null>(null);
  const notifications = data.notifications ?? [];
  const open = async (notification: AppNotification) => {
    setError(null);
    setBusy(notification.id);
    try {
      let story =
        notification.entityType === "story"
          ? data.stories.find((item) => item.id === notification.entityId)
          : null;
      if (!story && notification.entityType === "story" && navigator.onLine) {
        const refreshed = await client.refreshResource({
          workspaceId: data.workspaceId,
          resourceId: notification.entityId,
          kind: "story",
        });
        acceptData(refreshed);
        story = refreshed.stories.find(
          (item) => item.id === notification.entityId,
        );
      }
      if (!notification.readAt && navigator.onLine)
        try {
          acceptData(
            await client.markNotificationRead(
              data.workspaceId,
              notification.id,
              notification.createdAt,
            ),
          );
        } catch (reason) {
          setError(
            `Opened notification; read status will need a connection. ${errorMessage(reason)}`,
          );
        }
      if (story) onStory(story);
      else
        await client.openWorkspacePage(
          data.workspaceId,
          `/notifications/${encodeURIComponent(notification.id)}`,
        );
    } catch (reason) {
      setError(errorMessage(reason));
    } finally {
      setBusy(null);
    }
  };
  return (
    <section className="collection">
      <div className="page-heading">
        <div>
          <span className="eyebrow">Stay close to what matters</span>
          <h1>Inbox</h1>
          <p>
            {data.unreadNotificationCount ??
              notifications.filter((item) => !item.readAt).length}{" "}
            unread notifications · recent activity available offline
          </p>
        </div>
      </div>
      <ErrorNotice message={error} />
      {!notifications.length ? (
        <Empty title="A little quiet">
          Notifications will appear here as your team’s work moves forward.
        </Empty>
      ) : (
        <div className="inbox-list">
          {notifications.map((notification) => (
            <button
              className="notification-row"
              key={notification.id}
              disabled={Boolean(busy)}
              onClick={() => {
                void open(notification);
              }}
            >
              <span
                className={
                  notification.readAt ? "notification-read" : "pending-dot"
                }
              />
              <div>
                <strong>{notification.title}</strong>
                <p>{notificationMessage(notification)}</p>
                <time>{formatTime(notification.createdAt)}</time>
              </div>
            </button>
          ))}
        </div>
      )}
      <p className="footnote">
        Desktop alerts follow your notification preferences while FortyOne is
        running. Online-only destinations open in your browser.
      </p>
    </section>
  );
}
