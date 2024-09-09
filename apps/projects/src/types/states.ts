export type StateCategory =
  | "backlog"
  | "unstarted"
  | "started"
  | "paused"
  | "completed"
  | "cancelled";

export type State = {
  id: string;
  name: string;
  color: string;
  category: StateCategory;
  orderIndex: number;
  teamId: string;
  workspaceId: string;
  createdAt: string;
  updatedAt: string;
};
