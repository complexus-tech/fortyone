export type IssueStatus =
  | "Backlog"
  | "Todo"
  | "In Progress"
  | "Testing"
  | "Done"
  | "Paused"
  | "Duplicate"
  | "Canceled";

export type IssuePriority =
  | "No Priority"
  | "Urgent"
  | "High"
  | "Medium"
  | "Low";

export type Issue = {
  id: number;
  title: string;
  description?: string;
  status?: IssueStatus;
  priority?: IssuePriority;
};
