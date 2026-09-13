import { get } from "@/lib/http";
import type { ApiResponse } from "@/types";
import type { Sprint } from "../types";

export const getSprints = async (signal?: AbortSignal) => {
  const response = await get<ApiResponse<Sprint[]>>("sprints", { signal });
  return response.data ?? [];
};

export const getTeamSprints = async (teamId: string, signal?: AbortSignal) => {
  if (!teamId) return [];
  const response = await get<ApiResponse<Sprint[]>>(
    `sprints?teamId=${teamId}`,
    { signal },
  );
  return response.data ?? [];
};

export const getSprint = async (sprintId: string, signal?: AbortSignal) => {
  const response = await get<ApiResponse<Sprint>>(`sprints/${sprintId}`, {
    signal,
  });
  return response.data;
};
