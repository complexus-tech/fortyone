import { parseAsIsoDate, useQueryStates } from "nuqs";
import { formatISO } from "date-fns";
import { getDefaultDateRange } from "@/modules/analytics/components/filters/types";

export const useSummaryDateFilters = () => {
  const defaultDates = getDefaultDateRange();
  const [filters] = useQueryStates(
    {
      startDate: parseAsIsoDate.withDefault(defaultDates.startDate),
      endDate: parseAsIsoDate.withDefault(defaultDates.endDate),
    },
    {
      clearOnDefault: true,
    },
  );

  return {
    startDate: formatISO(filters.startDate, {
      representation: "date",
    }),
    endDate: formatISO(filters.endDate, {
      representation: "date",
    }),
  };
};
