export type StoryStatus =
  | "Backlog"
  | "Todo"
  | "In Progress"
  | "Testing"
  | "Done"
  | "Paused"
  | "Duplicate"
  | "Canceled";

export type StoryPriority =
  | "No Priority"
  | "Urgent"
  | "High"
  | "Medium"
  | "Low";

export type Story = {
  id: number;
  title: string;
  description?: string;
  status?: StoryStatus;
  priority?: StoryPriority;
};
