import { useQuery } from "@tanstack/react-query";
import { useSession } from "next-auth/react";
import { storyKeys } from "../constants";
import { getStories } from "../queries/get-stories";

export const useObjectiveStories = (objectiveId: string) => {
  const { data: session } = useSession();
  return useQuery({
    queryKey: storyKeys.objective(objectiveId),
    queryFn: () => getStories(session!, { objectiveId }),
  });
};
