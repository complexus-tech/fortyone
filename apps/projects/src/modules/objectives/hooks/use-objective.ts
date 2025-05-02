import { useQuery } from "@tanstack/react-query";
import { useSession } from "next-auth/react";
import { objectiveKeys } from "../constants";
import { getObjective } from "../queries/get-objective";

export const useObjective = (objectiveId: string) => {
  const { data: session } = useSession();
  return useQuery({
    queryKey: objectiveKeys.objective(objectiveId),
    queryFn: () => getObjective(objectiveId, session!),
  });
};
