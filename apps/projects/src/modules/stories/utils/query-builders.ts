import { stringify } from "qs";
import type { GroupedStoryParams, GroupStoryParams } from "../types";

export const buildGroupedStoriesQuery = (params: GroupedStoryParams) => {
  const query = stringify(params, {
    skipNulls: true,
    addQueryPrefix: true,
    encodeValuesOnly: true,
  });
  return query;
};

export const buildGroupStoriesQuery = (params: GroupStoryParams) => {
  const query = stringify(params, {
    skipNulls: true,
    addQueryPrefix: true,
    encodeValuesOnly: true,
  });
  return query;
};

export const getStoriesUrl = (type: "grouped" | "group") => `stories/${type}`;

export const buildQueryKey = (
  parts: (string | number | boolean | undefined | null)[],
) => parts.filter(Boolean).join("-");
