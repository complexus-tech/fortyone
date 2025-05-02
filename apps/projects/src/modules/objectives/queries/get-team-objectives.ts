import type { Session } from "next-auth";
import { get } from "@/lib/http";
import type { ApiResponse } from "@/types";
import type { Objective } from "../types";

export const getTeamObjectives = async (teamId: string, session: Session) => {
  const objectives = await get<ApiResponse<Objective[]>>(
    `objectives?teamId=${teamId}`,
    session,
  );
  return objectives.data ?? [];
};
