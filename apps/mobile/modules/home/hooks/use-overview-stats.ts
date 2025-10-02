import { useQuery } from "@tanstack/react-query";
import { getOverviewStats } from "../queries/get-overview-stats";
import { homeKeys } from "@/constants/keys";

export const useOverviewStats = () => {
  return useQuery({
    queryKey: homeKeys.overview(),
    queryFn: getOverviewStats,
  });
};
