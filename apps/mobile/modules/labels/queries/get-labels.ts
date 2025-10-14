import { get } from "@/lib/http/fetch";
import type { ApiResponse, Label } from "@/types";

export const getLabels = async (params: { teamId?: string } = {}) => {
  const query = new URLSearchParams();
  if (params.teamId) {
    query.append("teamId", params.teamId);
  }

  const queryString = query.toString();
  const url = queryString ? `labels?${queryString}` : "labels";

  const response = await get<ApiResponse<Label[]>>(url);
  return response.data!;
};
