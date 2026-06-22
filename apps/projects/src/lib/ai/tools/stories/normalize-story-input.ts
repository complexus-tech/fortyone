import { normalizeOptionalString } from "../normalize-input";

type StoryInput = {
  title: string;
  description?: string;
  descriptionHTML?: string;
  teamId: string;
  statusId: string;
  assigneeId?: string;
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
    title: story.title,
    teamId: story.teamId,
    statusId: story.statusId,
    priority: story.priority,
    estimateValue: story.estimateValue,
    labelIds: story.labelIds,
    description: normalizeOptionalString(story.description),
    descriptionHTML: normalizeOptionalString(story.descriptionHTML),
    assigneeId: normalizeOptionalString(story.assigneeId),
    sprintId: normalizeOptionalString(story.sprintId),
    objectiveId: normalizeOptionalString(story.objectiveId),
    parentId: normalizeOptionalString(story.parentId),
    startDate: normalizeOptionalString(story.startDate),
    endDate: normalizeOptionalString(story.endDate),
  };
};
