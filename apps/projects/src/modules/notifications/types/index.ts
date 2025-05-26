export type AppNotification = {
  id: string;
  recipientId: string;
  workspaceId: string;
  type: "story_update" | "story_comment";
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

export type NotificationType =
  | "story_update"
  | "objective_update"
  | "comment_reply"
  | "mention"
  | "key_result_update"
  | "story_comment";

export type UpdateNotificationPreferences = {
  emailEnabled?: boolean;
  inAppEnabled?: boolean;
};
