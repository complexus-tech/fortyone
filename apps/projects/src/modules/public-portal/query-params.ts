import type {
  PublicPortalFilters,
  PublicPortalSort,
  PublicRequestStatus,
} from "./types";

type SearchParamsRecord = Record<string, string | string[] | undefined>;
type SearchParamsSource = URLSearchParams | SearchParamsRecord;

const isSearchParams = (value: SearchParamsSource): value is URLSearchParams =>
  value instanceof URLSearchParams;

const getParam = (params: SearchParamsSource, key: string) => {
  if (isSearchParams(params)) return params.get(key) ?? undefined;

  const value = params[key];
  return Array.isArray(value) ? value[0] : value;
};

export const isPublicPortalSort = (value?: string): value is PublicPortalSort =>
  value === "top" || value === "newest" || value === "oldest";

export const isPublicRequestStatus = (
  value?: string,
): value is PublicRequestStatus =>
  value === "pending" ||
  value === "reviewing" ||
  value === "planned" ||
  value === "in_progress" ||
  value === "completed" ||
  value === "closed";

export const parsePublicPortalFilters = (
  params: SearchParamsSource,
): PublicPortalFilters => {
  const sort = getParam(params, "sort");
  const status = getParam(params, "status");

  return {
    boardId: getParam(params, "boardId") || undefined,
    search: getParam(params, "search")?.trim() ?? "",
    sort: isPublicPortalSort(sort) ? sort : "top",
    status: isPublicRequestStatus(status) ? status : undefined,
  };
};

export const toPublicPortalSearchParams = (filters: PublicPortalFilters) => {
  const params = new URLSearchParams({ sort: filters.sort });

  if (filters.boardId) params.set("boardId", filters.boardId);
  if (filters.search.trim()) params.set("search", filters.search.trim());
  if (filters.status) params.set("status", filters.status);

  return params;
};
