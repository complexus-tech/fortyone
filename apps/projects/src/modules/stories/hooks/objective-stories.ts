import { useQuery } from "@tanstack/react-query";
import { useSession } from "next-auth/react";
import { useWorkspacePath } from "@/hooks";
import { storyKeys } from "../constants";
import { getStories } from "../queries/get-stories";

export const useObjectiveStories = (objectiveId: string) => {
  const { data: session } = useSession();
  const { workspaceSlug } = useWorkspacePath();

  return useQuery({
    queryKey: storyKeys.objective(workspaceSlug, objectiveId),
    queryFn: () =>
      getStories({ session: session!, workspaceSlug }, { objectiveId }),
  });
};
