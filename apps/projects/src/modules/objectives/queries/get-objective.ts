import type { Session } from "next-auth";
import { get } from "@/lib/http";
import type { ApiResponse } from "@/types";
import type { Objective } from "../types";

export const getObjective = async (objectiveId: string, session: Session) => {
  const objective = await get<ApiResponse<Objective>>(
    `objectives/${objectiveId}`,
    session,
  );
  return objective.data;
};
