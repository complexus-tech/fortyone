import { get } from "@/lib/http";
import type { ApiResponse } from "@/types";
import type { Sprint } from "../types";

export const getSprints = async () => {
  const response = await get<ApiResponse<Sprint[]>>("sprints");
  return response.data ?? [];
};

export const getTeamSprints = async (teamId: string) => {
  if (!teamId) return [];
  const response = await get<ApiResponse<Sprint[]>>(`sprints?teamId=${teamId}`);
  return response.data ?? [];
};
