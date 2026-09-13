import { get } from "@/lib/http";
import type { ApiResponse } from "@/types";
import type { Status } from "@/types/statuses";

export const getStatuses = async (signal?: AbortSignal) => {
  const response = await get<ApiResponse<Status[]>>("states", { signal });
  return response.data ?? [];
};

export const getTeamStatuses = async (teamId: string, signal?: AbortSignal) => {
  if (!teamId) return [];
  const response = await get<ApiResponse<Status[]>>(`states?teamId=${teamId}`, {
    signal,
  });
  return response.data ?? [];
};
