import { get } from "@/lib/http";
import type { ApiResponse } from "@/types";
import type { Objective } from "../types";

export const getObjective = async (objectiveId: string) => {
  const objective = await get<ApiResponse<Objective>>(
    `objectives/${objectiveId}`,
  );
  return objective.data;
};
