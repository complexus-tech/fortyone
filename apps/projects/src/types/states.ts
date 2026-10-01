export type StateCategory =
  | "backlog"
  | "unstarted"
  | "started"
  | "paused"
  | "completed"
  | "cancelled";

export type State = {
  /** Advisory limit; a missing value preserves compatibility with older APIs. */
  wipLimit?: number | null;
  activeCount?: number;
  id: string;
  name: string;
  color: string;
  category: StateCategory;
  isDefault: boolean;
  orderIndex: number;
  teamId: string;
  workspaceId: string;
  createdAt: string;
  updatedAt: string;
};
