export type AppNotification = {
  id: string;
  recipientId: string;
  workspaceId: string;
  type: string;
  entityType:
    | "story"
    | "comment"
    | "objective"
    | "key_result"
    | "strategy"
    | "feedback"
    | "sprint";
  entityId: string;
  actorId: string;
  actor?: {
    id: string;
    username: string;
    fullName: string;
    avatarUrl: string;
    isActive: boolean;
    isSystem: boolean;
  };
  title: string;
  message: {
    template: string;
    variables: Record<string, { value: string; type?: string }> | null;
  };
  createdAt: string;
  readAt: string | null;
};

type NotificationChannel = {
  email: boolean;
  inApp: boolean;
  push: boolean;
};

export type NotificationPreferences = {
  id: string;
  userId: string;
  workspaceId: string;
  preferences: Partial<Record<NotificationType, NotificationChannel>>;
  createdAt: string;
  updatedAt: string;
};

export type NotificationType =
  | "story_update"
  | "objective_update"
  | "comment_reply"
  | "mention"
  | "key_result_update"
  | "story_comment"
  | "reminders"
  | "weekly_digest"
  | "strategy_update";

export type UpdateNotificationPreferences = {
  emailEnabled?: boolean;
  inAppEnabled?: boolean;
  pushEnabled?: boolean;
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
