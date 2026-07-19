import { useQuery } from "@tanstack/react-query";
import type { User } from "@/types";
import { userKeys } from "@/constants/keys";
import { DURATION_FROM_MILLISECONDS } from "@/constants/time";
import { getProfile } from "../queries/users/profile";

export const useProfile = (initialData?: User) => {
  return useQuery({
    queryKey: userKeys.profile(),
    queryFn: () => getProfile(),
    initialData,
    staleTime: DURATION_FROM_MILLISECONDS.MINUTE * 10,
  });
};
