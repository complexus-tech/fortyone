import {
  useQueryStates,
  parseAsIsoDateTime,
  parseAsArrayOf,
  parseAsString,
} from "nuqs";
import { formatISO } from "date-fns";
import type { AnalyticsFilters } from "@/modules/analytics/types";
import { getDefaultDateRange } from "../components/filters/types";

export const useAppliedFilters = () => {
  const [filters] = useQueryStates({
    startDate: parseAsIsoDateTime,
    endDate: parseAsIsoDateTime,
    teamIds: parseAsArrayOf(parseAsString),
    sprintIds: parseAsArrayOf(parseAsString),
    objectiveIds: parseAsArrayOf(parseAsString),
  });

  const defaultDates = getDefaultDateRange();
  const analyticsFilters: AnalyticsFilters = {
    startDate: formatISO(filters.startDate ?? defaultDates.startDate, {
      representation: "date",
    }),
    endDate: formatISO(filters.endDate ?? defaultDates.endDate, {
      representation: "date",
    }),
    teamIds: filters.teamIds?.length ? filters.teamIds : undefined,
    sprintIds: filters.sprintIds?.length ? filters.sprintIds : undefined,
    objectiveIds: filters.objectiveIds?.length
      ? filters.objectiveIds
      : undefined,
  };

  return analyticsFilters;
};
