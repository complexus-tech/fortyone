import type { Story, StoryPriority } from "@/modules/stories/types";

export type StoryAssociationType = "related" | "blocking" | "duplicate";

export type StoryAssociation = {
  id: string;
  fromStoryId: string;
  toStoryId: string;
  type: StoryAssociationType;
  story: Story;
};

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
  keyResultId: string | null;
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
  completedAt: string | null;
  archivedAt: string | null;
  subStories: Story[];
  labels: string[] | null;
  associations: StoryAssociation[];
};

export type NewStory = Partial<DetailedStory>;

export type StoryAttachment = {
  id: string;
  filename: string;
  size: number;
  mimeType: string;
  url: string;
  createdAt: string;
  uploadedBy: string;
};
