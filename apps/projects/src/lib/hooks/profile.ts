import { useQuery } from "@tanstack/react-query";
import { useSession } from "next-auth/react";
import { userKeys } from "@/constants/keys";
import { DURATION_FROM_MILLISECONDS } from "@/constants/time";
import { getProfile } from "../queries/users/profile";

export const useProfile = () => {
  const { data: session } = useSession();
  return useQuery({
    queryKey: userKeys.profile(),
    queryFn: () => getProfile(session!),
    staleTime: DURATION_FROM_MILLISECONDS.MINUTE * 10,
  });
};
