import { useQuery } from "@tanstack/react-query";
import { getTeams } from "../queries/get-teams";
import { teamKeys } from "@/constants/keys";

export const useTeams = () => {
  return useQuery({
    queryKey: teamKeys.lists(),
    queryFn: getTeams,
  });
};
