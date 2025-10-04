import { get } from "@/lib/http";
import type { ApiResponse } from "@/types";
import type { Objective } from "../types";

export const getObjective = async (objectiveId: string) => {
  const response = await get<ApiResponse<Objective>>(
    `objectives/${objectiveId}`
  );
  return response.data;
};
