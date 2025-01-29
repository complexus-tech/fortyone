import { get } from "@/lib/http";
import type { ApiResponse } from "@/types";
import { DURATION_FROM_SECONDS } from "@/constants/time";
import type { Objective } from "../types";
import { objectiveTags } from "../constants";

export const getObjective = async (objectiveId: string) => {
  const objective = await get<ApiResponse<Objective>>(
    `objectives/${objectiveId}`,
    {
      next: {
        revalidate: DURATION_FROM_SECONDS.MINUTE * 10,
        tags: [objectiveTags.objective(objectiveId)],
      },
    },
  );
  return objective.data;
};
