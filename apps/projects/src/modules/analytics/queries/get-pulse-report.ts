import { stringify } from "qs";
import { get, type WorkspaceCtx } from "@/lib/http";
import type { ApiResponse } from "@/types";
import type { AnalyticsFilters, PulseReport } from "../types";

const PULSE_REPORT_TIMEOUT_MS = 45_000;

export const getPulseReport = async (
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

  const report = await get<ApiResponse<PulseReport>>(
    `analytics/pulse${query}`,
    ctx,
    { timeout: PULSE_REPORT_TIMEOUT_MS },
  );
  return report.data!;
};
