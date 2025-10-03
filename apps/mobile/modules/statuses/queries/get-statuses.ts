import { get } from "@/lib/http";
import type { ApiResponse } from "@/types";
import type { Status } from "@/types/statuses";

export const getStatuses = async () => {
  const response = await get<ApiResponse<Status[]>>("states");
  return response.data ?? [];
};

export const getTeamStatuses = async (teamId: string) => {
  if (!teamId) return [];
  const response = await get<ApiResponse<Status[]>>(`states?teamId=${teamId}`);
  return response.data ?? [];
};
