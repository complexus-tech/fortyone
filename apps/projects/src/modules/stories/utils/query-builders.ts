import { stringify } from "qs";
import type { GroupedStoryParams, GroupStoryParams } from "../types";

type StoryQueryParams = GroupedStoryParams | GroupStoryParams;

const normalizeStoryQueryParams = (
  params: StoryQueryParams,
): Record<string, unknown> => {
  const normalized: Record<string, unknown> = {};

  for (const [key, value] of Object.entries(params)) {
    if (Array.isArray(value)) {
      const normalizedValues = value
        .map((item) => (typeof item === "string" ? item.trim() : item))
        .filter((item) => typeof item !== "string" || item.length > 0);

      if (normalizedValues.length === 0) continue;

      normalized[key] = normalizedValues;
      continue;
    }

    if (typeof value === "string") {
      const trimmedValue = value.trim();

      if (trimmedValue.length === 0) {
        continue;
      }

      normalized[key] = trimmedValue;
      continue;
    }

    normalized[key] = value;
  }

  return normalized;
};

export const buildGroupedStoriesQuery = (params: GroupedStoryParams) => {
  const query = stringify(normalizeStoryQueryParams(params), {
    skipNulls: true,
    addQueryPrefix: true,
    encodeValuesOnly: true,
    arrayFormat: "comma",
  });
  return query;
};

export const buildGroupStoriesQuery = (params: GroupStoryParams) => {
  const query = stringify(normalizeStoryQueryParams(params), {
    skipNulls: true,
    addQueryPrefix: true,
    encodeValuesOnly: true,
    arrayFormat: "comma",
  });
  return query;
};

export const getStoriesUrl = (type: "grouped" | "group") => `stories/${type}`;
