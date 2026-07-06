import { normalizeOptionalString } from "../normalize-input";

const ALLOWED_ESTIMATE_VALUES = new Set([1, 2, 3, 5, 8]);

type StoryInput = {
  title: string;
  description?: string | null;
  descriptionHTML?: string | null;
  teamId: string;
  statusId: string;
  assigneeId?: string | null;
  priority: "No Priority" | "Low" | "Medium" | "High" | "Urgent";
  estimateValue?: number | null;
  sprintId?: string | null;
  objectiveId?: string | null;
  labelIds?: string[];
  parentId?: string | null;
  startDate?: string | null;
  endDate?: string | null;
};

type NormalizedStoryInput = {
  title: string;
  teamId: string;
  statusId: string;
  priority: StoryInput["priority"];
  estimateValue?: number;
  labelIds?: string[];
  description?: string;
  descriptionHTML?: string;
  assigneeId?: string;
  sprintId?: string;
  objectiveId?: string;
  parentId?: string;
  startDate?: string;
  endDate?: string;
};

const isPlaceholderValue = (value: string) => {
  const trimmed = value.trim();
  return trimmed.startsWith("[") && trimmed.endsWith("]");
};

const normalizeRequiredId = (value: string, fieldName: string) => {
  const normalized = normalizeOptionalString(value);
  if (!normalized || isPlaceholderValue(normalized)) {
    throw new Error(
      `${fieldName} must be resolved to a real ID before creating a story.`,
    );
  }
  return normalized;
};

const normalizeOptionalId = (value?: string | null) => {
  const normalized = normalizeOptionalString(value);
  if (!normalized || isPlaceholderValue(normalized)) {
    return undefined;
  }
  return normalized;
};

const normalizeEstimateValue = (value?: number | null) => {
  if (value == null || value === 0) {
    return undefined;
  }

  if (!ALLOWED_ESTIMATE_VALUES.has(value)) {
    throw new Error("estimateValue must be one of 1, 2, 3, 5, or 8.");
  }

  return value;
};

const setIfDefined = <Key extends keyof NormalizedStoryInput>(
  payload: NormalizedStoryInput,
  key: Key,
  value: NormalizedStoryInput[Key] | undefined,
) => {
  if (value !== undefined) {
    payload[key] = value;
  }
};

export const normalizeStoryInput = <T extends StoryInput>(story: T) => {
  const payload: NormalizedStoryInput = {
    title: story.title,
    teamId: normalizeRequiredId(story.teamId, "teamId"),
    statusId: normalizeRequiredId(story.statusId, "statusId"),
    priority: story.priority,
  };

  setIfDefined(
    payload,
    "estimateValue",
    normalizeEstimateValue(story.estimateValue),
  );
  setIfDefined(payload, "labelIds", story.labelIds);
  setIfDefined(
    payload,
    "description",
    normalizeOptionalString(story.description),
  );
  setIfDefined(
    payload,
    "descriptionHTML",
    normalizeOptionalString(story.descriptionHTML),
  );
  setIfDefined(payload, "assigneeId", normalizeOptionalId(story.assigneeId));
  setIfDefined(payload, "sprintId", normalizeOptionalId(story.sprintId));
  setIfDefined(payload, "objectiveId", normalizeOptionalId(story.objectiveId));
  setIfDefined(payload, "parentId", normalizeOptionalId(story.parentId));
  setIfDefined(payload, "startDate", normalizeOptionalId(story.startDate));
  setIfDefined(payload, "endDate", normalizeOptionalId(story.endDate));

  return payload;
};
