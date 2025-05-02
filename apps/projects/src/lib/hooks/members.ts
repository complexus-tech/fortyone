import { useQuery } from "@tanstack/react-query";
import { useSession } from "next-auth/react";
import { getMembers } from "@/lib/queries/members/get-members";
import { memberKeys } from "@/constants/keys";
import { DURATION_FROM_MILLISECONDS } from "@/constants/time";

export const useMembers = () => {
  const { data: session } = useSession();
  return useQuery({
    queryKey: memberKeys.lists(),
    queryFn: () => getMembers(session!),
    staleTime: DURATION_FROM_MILLISECONDS.MINUTE * 10,
  });
};
