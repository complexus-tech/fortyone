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
  sequenceId: number;
};

export type DetailedStory = {
  id: string;
  sequenceId: number;
  title: string;
  description: string;
  descriptionHTML: string;
  parentId: string;
  objectiveId: string;
  statusId: string;
  assigneeId: string;
  blockedById: string;
  blockingId: string;
  relatedId: string;
  reporterId: string;
  priority: StoryPriority;
  sprintId: string;
  startDate: string;
  endDate: string;
  createdAt: string;
  updatedAt: string;
  deletedAt: string;
};
