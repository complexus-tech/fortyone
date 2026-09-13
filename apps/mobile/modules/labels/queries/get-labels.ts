import { get } from "@/lib/http/fetch";
import type { ApiResponse, Label } from "@/types";

export const getLabels = async (
  params: { teamId?: string } = {},
  signal?: AbortSignal,
) => {
  const query = new URLSearchParams();
  if (params.teamId) {
    query.append("teamId", params.teamId);
  }

  const queryString = query.toString();
  const url = queryString ? `labels?${queryString}` : "labels";

  const response = await get<ApiResponse<Label[]>>(url, { signal });
  return response.data!;
};
