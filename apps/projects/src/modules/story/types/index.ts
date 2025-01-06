import { Story, StoryPriority } from "@/modules/stories/types";

export type DetailedStory = {
  id: string;
  sequenceId: number;
  title: string;
  description: string;
  descriptionHTML: string;
  parentId: string;
  teamId: string;
  workspaceId: string;
  objectiveId: string | null;
  statusId: string;
  assigneeId: string | null;
  blockedById: string | null;
  blockingId: string | null;
  relatedId: string | null;
  reporterId: string;
  priority: StoryPriority;
  sprintId: string | null;
  epicId: string | null;
  startDate: string | null;
  endDate: string | null;
  createdAt: string;
  updatedAt: string;
  deletedAt: string | null;
  subStories: Story[];
  labels: string[];
};

export type NewStory = Partial<DetailedStory>;
