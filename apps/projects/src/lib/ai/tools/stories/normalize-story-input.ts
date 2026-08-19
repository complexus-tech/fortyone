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
  autoSchedulingEnabled?: boolean;
  sprintId?: string | null;
  objectiveId?: string | null;
  keyResultId?: string | null;
  labelIds?: string[] | null;
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
  autoSchedulingEnabled?: boolean;
  labelIds?: string[];
  description?: string;
  descriptionHTML?: string;
  assigneeId?: string;
  sprintId?: string;
  objectiveId?: string;
  keyResultId?: string;
  parentId?: string;
  startDate?: string;
  endDate?: string;
};

const isPlaceholderValue = (value: string) => {
  const trimmed = value.trim();
  return trimmed.startsWith("[") && trimmed.endsWith("]");
};

const UUID_PATTERN =
  /^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$/i;
const ISO_DATE_PATTERN = /^\d{4}-\d{2}-\d{2}$/;

const isValidUuid = (value: string) => UUID_PATTERN.test(value);

const normalizeRequiredString = (value: string, fieldName: string) => {
  const normalized = normalizeOptionalString(value);
  if (!normalized) {
    throw new Error(`${fieldName} is required.`);
  }
  return normalized;
};

export const normalizeRequiredStoryId = (value: string, fieldName: string) => {
  const normalized = normalizeOptionalString(value);
  if (!normalized || isPlaceholderValue(normalized)) {
    throw new Error(
      `${fieldName} must be resolved to a real ID before creating a story.`,
    );
  }
  if (!isValidUuid(normalized)) {
    throw new Error(`${fieldName} must be a valid UUID.`);
  }
  return normalized;
};

export const normalizeOptionalStoryId = (
  value: string | null | undefined,
  fieldName: string,
) => {
  const normalized = normalizeOptionalString(value);
  if (!normalized || isPlaceholderValue(normalized)) {
    return undefined;
  }
  if (!isValidUuid(normalized)) {
    throw new Error(`${fieldName} must be a valid UUID.`);
  }
  return normalized;
};

const normalizeOptionalIds = (values?: string[] | null) => {
  if (values === null || values === undefined) return undefined;

  const normalizedIds = values.flatMap((value) => {
    const normalized = normalizeOptionalString(value);
    if (!normalized || isPlaceholderValue(normalized)) return [];
    if (!isValidUuid(normalized)) {
      throw new Error("labelIds must contain only valid UUIDs.");
    }
    return [normalized];
  });

  return normalizedIds.length > 0 ? [...new Set(normalizedIds)] : undefined;
};

const normalizeOptionalDate = (
  value: string | null | undefined,
  fieldName: "startDate" | "endDate",
) => {
  const normalized = normalizeOptionalString(value);
  if (!normalized || isPlaceholderValue(normalized)) return undefined;

  const dateValue = normalized.slice(0, 10);
  const timeValue = normalized.slice(10);
  if (
    !ISO_DATE_PATTERN.test(dateValue) ||
    (timeValue && !timeValue.startsWith("T"))
  ) {
    throw new Error(`${fieldName} must be an ISO date in YYYY-MM-DD format.`);
  }

  const [year, month, day] = dateValue.split("-").map(Number);
  const date = new Date(Date.UTC(year, month - 1, day));
  if (
    date.getUTCFullYear() !== year ||
    date.getUTCMonth() !== month - 1 ||
    date.getUTCDate() !== day
  ) {
    throw new Error(`${fieldName} must be a valid calendar date.`);
  }

  return dateValue;
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
    title: normalizeRequiredString(story.title, "title"),
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
  setIfDefined(payload, "autoSchedulingEnabled", story.autoSchedulingEnabled);
  setIfDefined(payload, "labelIds", normalizeOptionalIds(story.labelIds));
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
  setIfDefined(
    payload,
    "assigneeId",
    normalizeOptionalStoryId(story.assigneeId, "assigneeId"),
  );
  setIfDefined(
    payload,
    "sprintId",
    normalizeOptionalStoryId(story.sprintId, "sprintId"),
  );
  setIfDefined(
    payload,
    "objectiveId",
    normalizeOptionalStoryId(story.objectiveId, "objectiveId"),
  );
  setIfDefined(
    payload,
    "keyResultId",
    normalizeOptionalStoryId(story.keyResultId, "keyResultId"),
  );
  setIfDefined(
    payload,
    "parentId",
    normalizeOptionalStoryId(story.parentId, "parentId"),
  );
  const startDate = normalizeOptionalDate(story.startDate, "startDate");
  const endDate = normalizeOptionalDate(story.endDate, "endDate");
  if (startDate && endDate && startDate > endDate) {
    throw new Error("endDate cannot be before startDate.");
  }
  setIfDefined(payload, "startDate", startDate);
  setIfDefined(payload, "endDate", endDate);

  return payload;
};
