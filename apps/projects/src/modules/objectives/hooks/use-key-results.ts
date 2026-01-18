import { useQuery } from "@tanstack/react-query";
import { useSession } from "next-auth/react";
import { useWorkspacePath } from "@/hooks";
import { DURATION_FROM_MILLISECONDS } from "@/constants/time";
import { objectiveKeys } from "../constants";
import { getKeyResults } from "../queries/get-key-results";

export const useKeyResults = (objectiveId: string) => {
  const { data: session } = useSession();
  const { workspaceSlug } = useWorkspacePath();
  return useQuery({
    queryKey: objectiveKeys.keyResults(workspaceSlug, objectiveId),
    queryFn: () => getKeyResults(objectiveId, { session: session!, workspaceSlug }),
    staleTime: DURATION_FROM_MILLISECONDS.MINUTE * 10,
  });
};
