import { normalizeOptionalString } from "../normalize-input";

type StoryInput = {
  title: string;
  description?: string;
  descriptionHTML?: string;
  teamId: string;
  statusId: string;
  assigneeId?: string;
  reporterId?: string;
  priority: "No Priority" | "Low" | "Medium" | "High" | "Urgent";
  estimateValue?: number;
  sprintId?: string;
  objectiveId?: string;
  labelIds?: string[];
  parentId?: string;
  startDate?: string;
  endDate?: string;
};

export const normalizeStoryInput = <T extends StoryInput>(story: T) => {
  return {
    ...story,
    description: normalizeOptionalString(story.description),
    descriptionHTML: normalizeOptionalString(story.descriptionHTML),
    assigneeId: normalizeOptionalString(story.assigneeId),
    reporterId: normalizeOptionalString(story.reporterId),
    sprintId: normalizeOptionalString(story.sprintId),
    objectiveId: normalizeOptionalString(story.objectiveId),
    parentId: normalizeOptionalString(story.parentId),
    startDate: normalizeOptionalString(story.startDate),
    endDate: normalizeOptionalString(story.endDate),
  };
};
