import { useQuery } from "@tanstack/react-query";
import { userKeys } from "@/constants/keys";
import { DURATION_FROM_MILLISECONDS } from "@/constants/time";
import { getProfile } from "../queries/users/profile";

export const useProfile = () => {
  return useQuery({
    queryKey: userKeys.profile(),
    queryFn: () => getProfile(),
    staleTime: DURATION_FROM_MILLISECONDS.MINUTE * 10,
  });
};
