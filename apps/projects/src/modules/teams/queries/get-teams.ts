import { get } from "@/lib/http";
import { Team } from "@/modules/teams/types";

export const getTeams = async () => {
  const teams = await get<Team[]>(`/teams`);
  return teams;
};
