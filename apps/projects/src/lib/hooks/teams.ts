import { useQuery } from "@tanstack/react-query";
import { getTeams } from "@/modules/teams/queries/get-teams";

export const useTeams = () => {
  return useQuery({
    queryKey: ["teams"],
    queryFn: getTeams,
  });
};
