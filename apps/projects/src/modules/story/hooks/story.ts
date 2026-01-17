import { useQuery } from "@tanstack/react-query";
import { useSession } from "next-auth/react";
import { useWorkspacePath } from "@/hooks";
import { storyKeys } from "@/modules/stories/constants";
import { DURATION_FROM_MILLISECONDS } from "@/constants/time";
import { getStory } from "../queries/get-story";

export const useStoryById = (id: string) => {
  const { data: session } = useSession();
  const { workspaceSlug } = useWorkspacePath();

  return useQuery({
    queryKey: storyKeys.detail(workspaceSlug, id),
    queryFn: () => getStory(id, { session: session!, workspaceSlug }),
    staleTime: Number(DURATION_FROM_MILLISECONDS.MINUTE),
  });
};
