import { get } from "@/lib/http";
import type { ApiResponse } from "@/types";
import type { Objective } from "../types";

export const getTeamObjectives = async (teamId: string) => {
  const objectives = await get<ApiResponse<Objective[]>>(
    `objectives?teamId=${teamId}`,
  );
  return objectives.data ?? [];
};
