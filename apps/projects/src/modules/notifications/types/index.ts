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
