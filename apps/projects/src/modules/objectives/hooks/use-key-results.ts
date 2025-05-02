import { useQuery } from "@tanstack/react-query";
import { useSession } from "next-auth/react";
import { DURATION_FROM_MILLISECONDS } from "@/constants/time";
import { objectiveKeys } from "../constants";
import { getKeyResults } from "../queries/get-key-results";

export const useKeyResults = (objectiveId: string) => {
  const { data: session } = useSession();
  return useQuery({
    queryKey: objectiveKeys.keyResults(objectiveId),
    queryFn: () => getKeyResults(objectiveId, session!),
    staleTime: DURATION_FROM_MILLISECONDS.MINUTE * 10,
  });
};
