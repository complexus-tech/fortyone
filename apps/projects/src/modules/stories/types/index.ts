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
  sprintId: string | null;
  objectiveId: string | null;
  keyResultId: string | null;
  teamId: string;
  workspaceId: string;
  assigneeId: string | null;
  reporterId: string;
  epicId: string | null;
  sequenceId: number;
  priority: StoryPriority;
  startDate: string | null;
  endDate: string | null;
  createdAt: string;
  updatedAt: string;
  labels: string[] | null;
  subStories: Story[];
};

export type StoryActivity = {
  id: string;
  storyId: string;
  userId: string;
  type: "update" | "create";
  field: string;
  currentValue: string;
  createdAt: string;
};
