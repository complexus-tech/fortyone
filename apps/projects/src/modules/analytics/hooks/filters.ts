import {
  useQueryStates,
  parseAsIsoDateTime,
  parseAsArrayOf,
  parseAsString,
} from "nuqs";
import { formatISO } from "date-fns";
import type { AnalyticsFilters } from "@/modules/analytics/types";

export const useAppliedFilters = () => {
  const [filters] = useQueryStates({
    startDate: parseAsIsoDateTime,
    endDate: parseAsIsoDateTime,
    teamIds: parseAsArrayOf(parseAsString),
    sprintIds: parseAsArrayOf(parseAsString),
    objectiveIds: parseAsArrayOf(parseAsString),
  });

  const analyticsFilters: AnalyticsFilters = {
    startDate: filters.startDate
      ? formatISO(filters.startDate, {
          representation: "date",
        })
      : undefined,
    endDate: filters.endDate
      ? formatISO(filters.endDate, {
          representation: "date",
        })
      : undefined,
    teamIds: filters.teamIds?.length ? filters.teamIds : undefined,
    sprintIds: filters.sprintIds?.length ? filters.sprintIds : undefined,
    objectiveIds: filters.objectiveIds?.length
      ? filters.objectiveIds
      : undefined,
  };

  return analyticsFilters;
};
