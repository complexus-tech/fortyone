import { stringify } from "qs";
import { get } from "@/lib/http";
import type { WorkspaceCtx } from "@/lib/http";
import type { ApiResponse, Label } from "@/types";

export const getLabels = async (
  ctx: WorkspaceCtx,
  params: {
    teamId?: string;
  } = {},
) => {
  const query = stringify(params, {
    skipNulls: true,
    addQueryPrefix: true,
    encodeValuesOnly: true,
  });
  const labels = await get<ApiResponse<Label[]>>(`labels${query}`, ctx);
  return labels.data!;
};
