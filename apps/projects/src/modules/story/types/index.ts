import { StoryPriority } from "@/modules/stories/types";

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

export type NewStory = {
  title: string;
  description: string;
  descriptionHTML: string;
  parentId?: string;
  objectiveId?: string;
  statusId?: string;
  assigneeId?: string;
  blockedById?: string;
  priority: StoryPriority;
  sprintId?: string;
  teamId: string;
  startDate?: string | null;
  endDate?: string | null;
};
