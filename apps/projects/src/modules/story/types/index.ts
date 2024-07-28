import { StoryPriority } from "@/modules/stories/types";

export type DetailedStory = {
  id: string;
  sequenceId: number;
  title: string;
  description: string;
  descriptionHTML: string;
  parentId: string;
  teamId: string;
  objectiveId: string;
  statusId: string;
  assigneeId: string;
  blockedById: string;
  blockingId: string;
  relatedId: string;
  reporterId: string;
  priority: StoryPriority;
  sprintId: string;
  startDate: string | null;
  endDate: string | null;
  createdAt: string;
  updatedAt: string;
  deletedAt: string;
};

export type NewStory = Partial<DetailedStory>;
