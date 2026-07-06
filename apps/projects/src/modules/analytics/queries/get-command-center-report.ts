import { stringify } from "qs";
import { get, type WorkspaceCtx } from "@/lib/http";
import type { ApiResponse } from "@/types";
import type { AnalyticsFilters, WorkspaceCommandCenterReport } from "../types";

const COMMAND_CENTER_TIMEOUT_MS = 60_000;

export const getCommandCenterReport = async (
  ctx: WorkspaceCtx,
  filters?: AnalyticsFilters,
) => {
  const query = filters
    ? stringify(filters, {
        skipNulls: true,
        addQueryPrefix: true,
        encodeValuesOnly: true,
        arrayFormat: "comma",
      })
    : "";

  const report = await get<ApiResponse<WorkspaceCommandCenterReport>>(
    `analytics/command-center${query}`,
    ctx,
    { timeout: COMMAND_CENTER_TIMEOUT_MS },
  );

  return report.data!;
};
