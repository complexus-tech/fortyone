export type AppNotification = {
  id: string;
  recipientId: string;
  workspaceId: string;
  type: "story" | "objective";
  entityType: string;
  entityId: string;
  actorId: string;
  title: string;
  description: string;
  createdAt: string;
  readAt: string | null;
};

type NotificationChannel = {
  email: boolean;
  inApp: boolean;
};

export type NotificationPreferences = {
  id: string;
  userId: string;
  workspaceId: string;
  preferences: {
    comment_reply: NotificationChannel;
    key_result_update: NotificationChannel;
    mention: NotificationChannel;
    objective_update: NotificationChannel;
    story_comment: NotificationChannel;
    story_update: NotificationChannel;
  };
  createdAt: string;
  updatedAt: string;
};

export type UpdateNotificationPreferences = {
  emailEnabled?: boolean;
  inAppEnabled?: boolean;
};
