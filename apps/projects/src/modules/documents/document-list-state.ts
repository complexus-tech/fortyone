import { DOCUMENTS_PAGE_SIZE } from "./constants";
import type {
  DocumentAccessFilter,
  DocumentListState,
  DocumentOwnerFilter,
  DocumentScope,
  DocumentSortDirection,
  DocumentSortField,
  DocumentUpdatedFilter,
  WorkspaceDocumentSummary,
} from "./types";

type SearchParamsReader = Pick<URLSearchParams, "get">;

const DAY_IN_MILLISECONDS = 24 * 60 * 60 * 1000;

const isAccessFilter = (value: string | null): value is DocumentAccessFilter =>
  value === "workspace" || value === "restricted" || value === "private";

const isOwnerFilter = (value: string | null): value is DocumentOwnerFilter =>
  value === "mine" || value === "others";

const isUpdatedFilter = (
  value: string | null,
): value is DocumentUpdatedFilter =>
  value === "today" || value === "7d" || value === "30d" || value === "90d";

const isSortField = (value: string | null): value is DocumentSortField =>
  value === "updated" || value === "title";

const isSortDirection = (
  value: string | null,
): value is DocumentSortDirection => value === "asc" || value === "desc";

const getPositiveInteger = (value: string | null) => {
  if (!value) return 1;
  const parsed = Number.parseInt(value, 10);
  return Number.isSafeInteger(parsed) && parsed > 0 ? parsed : 1;
};

export const getDocumentListState = (
  searchParams: SearchParamsReader,
): DocumentListState => {
  const rawAccess = searchParams.get("access");
  const rawDirection = searchParams.get("direction");
  const rawOwner = searchParams.get("owner");
  const rawSort = searchParams.get("sort");
  const rawUpdated = searchParams.get("updated");
  const sort = isSortField(rawSort) ? rawSort : "updated";
  const defaultDirection = sort === "title" ? "asc" : "desc";

  return {
    access: isAccessFilter(rawAccess) ? rawAccess : "all",
    direction: isSortDirection(rawDirection) ? rawDirection : defaultDirection,
    owner: isOwnerFilter(rawOwner) ? rawOwner : "all",
    page: getPositiveInteger(searchParams.get("page")),
    sort,
    updated: isUpdatedFilter(rawUpdated) ? rawUpdated : "all",
  };
};

const getUpdatedAfter = (filter: DocumentUpdatedFilter, now: Date) => {
  if (filter === "all") return null;
  if (filter === "today") {
    const startOfToday = new Date(now);
    startOfToday.setHours(0, 0, 0, 0);
    return startOfToday.getTime();
  }

  let days = 90;
  if (filter === "7d") days = 7;
  if (filter === "30d") days = 30;
  return now.getTime() - days * DAY_IN_MILLISECONDS;
};

const getTimestamp = (value: string) => {
  const timestamp = new Date(value).getTime();
  return Number.isNaN(timestamp) ? 0 : timestamp;
};

export const filterAndSortDocumentSummaries = ({
  currentUserId,
  documents,
  now = new Date(),
  scope,
  state,
}: {
  currentUserId?: string;
  documents: WorkspaceDocumentSummary[];
  now?: Date;
  scope: Exclude<DocumentScope, "templates">;
  state: DocumentListState;
}) => {
  const updatedAfter = getUpdatedAfter(state.updated, now);
  const owner = scope === "mine" ? "all" : state.owner;
  const filtered = documents.filter((document) => {
    if (state.access !== "all" && document.visibility !== state.access) {
      return false;
    }
    if (
      owner === "mine" &&
      (!currentUserId || document.createdBy !== currentUserId)
    ) {
      return false;
    }
    if (
      owner === "others" &&
      (!currentUserId || document.createdBy === currentUserId)
    ) {
      return false;
    }
    if (
      updatedAfter !== null &&
      getTimestamp(document.updatedAt) < updatedAfter
    ) {
      return false;
    }
    return true;
  });

  return [...filtered].sort((left, right) => {
    let comparison = 0;
    if (state.sort === "title") {
      comparison = left.title.localeCompare(right.title, undefined, {
        sensitivity: "base",
      });
    } else {
      const leftDate = getTimestamp(left.updatedAt);
      const rightDate = getTimestamp(right.updatedAt);
      comparison = leftDate - rightDate;
    }

    if (comparison === 0) comparison = left.id.localeCompare(right.id);
    return state.direction === "asc" ? comparison : -comparison;
  });
};

export const paginateDocumentSummaries = (
  documents: WorkspaceDocumentSummary[],
  requestedPage: number,
  pageSize = DOCUMENTS_PAGE_SIZE,
) => {
  const total = documents.length;
  const pageCount = Math.max(1, Math.ceil(total / pageSize));
  const page = Math.min(Math.max(1, requestedPage), pageCount);
  const offset = (page - 1) * pageSize;
  const items = documents.slice(offset, offset + pageSize);

  return {
    end: total === 0 ? 0 : offset + items.length,
    items,
    page,
    pageCount,
    start: total === 0 ? 0 : offset + 1,
    total,
  };
};
