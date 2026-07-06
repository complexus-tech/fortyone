export type PaginationInput = {
  page?: number;
  pageSize?: number;
};

export type PaginationMeta = {
  page: number;
  pageSize: number;
  totalCount: number;
  totalPages: number;
  hasMore: boolean;
  nextPage: number | null;
};

export type ActivityFilterInput = {
  userId?: string;
  fields?: string[];
  since?: string;
  until?: string;
};

export type ActivityLike = {
  userId?: string | null;
  field?: string | null;
  createdAt: string;
};

const DEFAULT_PAGE = 1;
const DEFAULT_PAGE_SIZE = 20;
const MAX_PAGE_SIZE = 100;

const toPositiveInteger = (value: number | undefined, fallback: number) => {
  if (!Number.isFinite(value) || value === undefined) {
    return fallback;
  }

  return Math.max(1, Math.trunc(value));
};

export const resolvePaginationInput = ({
  page,
  pageSize,
}: PaginationInput = {}) => ({
  page: toPositiveInteger(page, DEFAULT_PAGE),
  pageSize: Math.min(
    toPositiveInteger(pageSize, DEFAULT_PAGE_SIZE),
    MAX_PAGE_SIZE,
  ),
});

export const buildPaginationMeta = ({
  page,
  pageSize,
  totalCount,
}: {
  page: number;
  pageSize: number;
  totalCount: number;
}): PaginationMeta => {
  const totalPages = Math.max(1, Math.ceil(totalCount / pageSize));
  const hasMore = page < totalPages;

  return {
    page,
    pageSize,
    totalCount,
    totalPages,
    hasMore,
    nextPage: hasMore ? page + 1 : null,
  };
};

export const paginateRecords = <T>(
  records: T[],
  input: PaginationInput = {},
) => {
  const { page, pageSize } = resolvePaginationInput(input);
  const start = (page - 1) * pageSize;
  const pagedRecords = records.slice(start, start + pageSize);

  return {
    records: pagedRecords,
    pagination: buildPaginationMeta({
      page,
      pageSize,
      totalCount: records.length,
    }),
  };
};

export const requireToolConfirmation = (action: string) => ({
  success: false,
  needsConfirmation: true,
  message: `Please confirm before I ${action}.`,
});

export const filterActivityTimeline = <T extends ActivityLike>(
  activities: T[],
  filters: ActivityFilterInput = {},
) => {
  const fieldSet = new Set(
    filters.fields?.map((field) => field.trim()).filter(Boolean) ?? [],
  );
  const sinceTime = filters.since ? new Date(filters.since).getTime() : null;
  const untilTime = filters.until ? new Date(filters.until).getTime() : null;

  return activities.filter((activity) => {
    if (filters.userId && activity.userId !== filters.userId) {
      return false;
    }

    if (fieldSet.size > 0 && !fieldSet.has(activity.field ?? "")) {
      return false;
    }

    const createdAt = new Date(activity.createdAt).getTime();

    if (sinceTime !== null && createdAt < sinceTime) {
      return false;
    }

    if (untilTime !== null && createdAt > untilTime) {
      return false;
    }

    return true;
  });
};
