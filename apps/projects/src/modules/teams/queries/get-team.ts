import { get } from "@/lib/http";
import type { ApiResponse } from "@/types";
import type { Team } from "@/modules/teams/types";

export const getTeam = async (id: string): Promise<Team> => {
  const team = await get<ApiResponse<Team>>(`teams/${id}`);
  return team.data!;
};
