import { stringify } from "qs";
import type { Session } from "next-auth";
import { get } from "@/lib/http";
import type { ApiResponse } from "@/types";
import type { WorkspaceOverview, AnalyticsFilters } from "../types";

export const getWorkspaceOverview = async (
  filters?: AnalyticsFilters,
  session?: Session,
) => {
  if (!session) return null;

  const query = filters
    ? stringify(filters, {
        skipNulls: true,
        addQueryPrefix: true,
        encodeValuesOnly: true,
      })
    : "";

  const overview = await get<ApiResponse<WorkspaceOverview>>(
    `analytics/overview${query}`,
    session,
  );
  return overview.data!;
};
