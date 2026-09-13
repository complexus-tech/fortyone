import { get } from "@/lib/http";
import type { ApiResponse } from "@/types";
import type { Objective, ObjectiveStatus } from "../types";

export const getObjectives = async (signal?: AbortSignal) => {
  const response = await get<ApiResponse<Objective[]>>("objectives", {
    signal,
  });
  return response.data ?? [];
};

export const getTeamObjectives = async (
  teamId: string,
  signal?: AbortSignal,
) => {
  if (!teamId) return [];
  const response = await get<ApiResponse<Objective[]>>(
    `objectives?teamId=${teamId}`,
    { signal },
  );
  return response.data ?? [];
};

export const getObjectiveStatuses = async (signal?: AbortSignal) => {
  const response = await get<ApiResponse<ObjectiveStatus[]>>(
    "objective-statuses",
    { signal },
  );
  return response.data ?? [];
};

export const getObjective = async (
  objectiveId: string,
  signal?: AbortSignal,
) => {
  const response = await get<ApiResponse<Objective>>(
    `objectives/${objectiveId}`,
    { signal },
  );
  return response.data;
};
