import { useQuery } from "@tanstack/react-query";
import { useSession } from "next-auth/react";
import { objectiveKeys } from "../constants";
import { getObjectiveAnalytics } from "../queries/get-objective-analytics";

export const useObjectiveAnalytics = (objectiveId: string) => {
  const { data: session } = useSession();
  return useQuery({
    queryKey: objectiveKeys.analytics(objectiveId),
    queryFn: () => getObjectiveAnalytics(objectiveId, session!),
    enabled: Boolean(objectiveId),
  });
};
