import { get } from "@/lib/http";
import type { ApiResponse } from "@/types";
import type { Objective, ObjectiveStatus } from "../types";

export const getObjectives = async () => {
  const response = await get<ApiResponse<Objective[]>>("objectives");
  return response.data ?? [];
};

export const getTeamObjectives = async (teamId: string) => {
  if (!teamId) return [];
  const response = await get<ApiResponse<Objective[]>>(
    `objectives?teamId=${teamId}`
  );
  return response.data ?? [];
};

export const getObjectiveStatuses = async () => {
  const response =
    await get<ApiResponse<ObjectiveStatus[]>>("objective-statuses");
  return response.data ?? [];
};

export const getObjective = async (objectiveId: string) => {
  const response = await get<ApiResponse<Objective>>(
    `objectives/${objectiveId}`
  );
  return response.data;
};
