export type StoryPriority =
  | "No Priority"
  | "Urgent"
  | "High"
  | "Medium"
  | "Low";

export type Story = {
  id: string;
  title: string;
  description?: string;
  statusId: string;
  sprintId: string;
  objectiveId: string;
  teamId: string;
  workspaceId: string;
  epicId: string;
  sequenceId: string;
  priority: StoryPriority;
  startDate: string;
  endDate: string;
  createdAt: string;
  updatedAt: string;
};
