import type { KeyResultFilters } from "./types";

export const getActiveKeyResultFilterCount = (filters: KeyResultFilters) =>
  Number(Boolean(filters.teamIds?.length)) +
  Number(Boolean(filters.objectiveIds?.length)) +
  Number(Boolean(filters.measurementTypes?.length)) +
  Number(Boolean(filters.leadIds?.length)) +
  Number(Boolean(filters.endDateAfter || filters.endDateBefore));
