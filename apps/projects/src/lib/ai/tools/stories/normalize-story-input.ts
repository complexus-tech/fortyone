import { isEstimateValue } from "@/lib/estimate";
import { MAX_TIME_NEEDED_MINUTES } from "@/lib/time-needed";
import { normalizeOptionalString } from "../normalize-input";

type StoryInput = {
  title: string;
  description?: string | null;
  descriptionHTML?: string | null;
  teamId: string;
  statusId: string;
  assigneeId?: string | null;
  priority: "No Priority" | "Low" | "Medium" | "High" | "Urgent";
  estimateValue?: number | null;
  estimatedDurationMinutes?: number | null;
  minimumFocusBlockMinutes?: number | null;
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
  estimatedDurationMinutes?: number;
  minimumFocusBlockMinutes?: number;
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

export const normalizeRequiredStoryId = (value: string, fieldName: string) => {
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
  if (value === null || value === undefined || value === 0) {
    return undefined;
  }

  if (!isEstimateValue(value)) {
    throw new Error("estimateValue must be one of 1, 2, 3, 5, or 8.");
  }

  return value;
};

const normalizePositiveMinutes = (
  value: number | null | undefined,
  fieldName: string,
) => {
  if (value === null || value === undefined) return undefined;
  if (
    !Number.isInteger(value) ||
    value <= 0 ||
    value > MAX_TIME_NEEDED_MINUTES
  ) {
    throw new Error(`${fieldName} must be a positive whole number of minutes.`);
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
    teamId: normalizeRequiredStoryId(story.teamId, "teamId"),
    statusId: normalizeRequiredStoryId(story.statusId, "statusId"),
    priority: story.priority,
  };

  setIfDefined(
    payload,
    "estimateValue",
    normalizeEstimateValue(story.estimateValue),
  );
  const estimatedDurationMinutes = normalizePositiveMinutes(
    story.estimatedDurationMinutes,
    "estimatedDurationMinutes",
  );
  const minimumFocusBlockMinutes = normalizePositiveMinutes(
    story.minimumFocusBlockMinutes,
    "minimumFocusBlockMinutes",
  );
  if (
    minimumFocusBlockMinutes !== undefined &&
    (estimatedDurationMinutes === undefined ||
      minimumFocusBlockMinutes > estimatedDurationMinutes)
  ) {
    throw new Error(
      "minimumFocusBlockMinutes requires a duration and cannot exceed estimatedDurationMinutes.",
    );
  }
  setIfDefined(payload, "estimatedDurationMinutes", estimatedDurationMinutes);
  setIfDefined(payload, "minimumFocusBlockMinutes", minimumFocusBlockMinutes);
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
