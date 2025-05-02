import { useQuery } from "@tanstack/react-query";
import { useSession } from "next-auth/react";
import { sprintKeys } from "@/constants/keys";
import { DURATION_FROM_MILLISECONDS } from "@/constants/time";
import { getSprints } from "@/modules/sprints/queries/get-sprints";

export const useSprints = () => {
  const { data: session } = useSession();
  return useQuery({
    queryKey: sprintKeys.lists(),
    queryFn: () => getSprints(session!),
    staleTime: DURATION_FROM_MILLISECONDS.MINUTE * 10,
  });
};
