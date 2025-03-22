import { useQuery } from "@tanstack/react-query";
import { getMembers } from "@/lib/queries/members/get-members";
import { memberKeys } from "@/constants/keys";
import { DURATION_FROM_MILLISECONDS } from "@/constants/time";

export const useMembers = () => {
  return useQuery({
    queryKey: memberKeys.lists(),
    queryFn: () => getMembers(),
    staleTime: DURATION_FROM_MILLISECONDS.MINUTE * 10,
  });
};
