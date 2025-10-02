export type StatusCategory =
  | "backlog"
  | "unstarted"
  | "started"
  | "paused"
  | "completed"
  | "cancelled";

export type Status = {
  id: string;
  name: string;
  color: string;
  category: StatusCategory;
  isDefault: boolean;
  orderIndex: number;
  teamId: string;
  workspaceId: string;
  createdAt: string;
  updatedAt: string;
};
