export type AppNotification = {
  id: string;
  recipientId: string;
  workspaceId: string;
  type: "story_update" | "story_comment" | "mention";
  entityType: "story" | "objective";
  entityId: string;
  actorId: string;
  title: string;
  message: {
    template: string;
    variables: {
      actor: {
        value: string;
      };
      field: {
        value: string;
      };
      value: {
        value: string;
      };
    };
  };
  createdAt: string;
  readAt: string | null;
};

export type NotificationsPage = {
  notifications: AppNotification[];
  pagination: {
    page: number;
    pageSize: number;
    hasMore: boolean;
    nextPage: number;
  };
};

type NotificationChannel = {
  email: boolean;
  inApp: boolean;
};

export type NotificationType =
  | "story_update"
  | "objective_update"
  | "comment_reply"
  | "mention"
  | "key_result_update"
  | "story_comment"
  | "reminders"
  | "weekly_digest";

export type NotificationPreferences = {
  id: string;
  userId: string;
  workspaceId: string;
  preferences: Partial<Record<NotificationType, NotificationChannel>>;
  createdAt: string;
  updatedAt: string;
};

export type UpdateNotificationPreferences = {
  emailEnabled?: boolean;
  inAppEnabled?: boolean;
};
