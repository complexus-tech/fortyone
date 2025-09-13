import type { Session } from "next-auth";
import { get } from "@/lib/http";
import type { ApiResponse } from "@/types";
import type { Team } from "@/modules/teams/types";

export const getTeam = async (id: string, session: Session) => {
  const team = await get<ApiResponse<Team>>(`teams/${id}`, session);
  return team.data;
};
