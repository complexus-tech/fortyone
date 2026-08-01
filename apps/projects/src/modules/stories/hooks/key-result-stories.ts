import { useQuery } from "@tanstack/react-query";
import { useWorkspacePath } from "@/hooks";
import { useSession } from "@/lib/auth/client";
import { storyKeys } from "../constants";
import { getStories } from "../queries/get-stories";

export const useKeyResultStories = (keyResultId: string, enabled = true) => {
  const { data: session } = useSession();
  const { workspaceSlug } = useWorkspacePath();

  return useQuery({
    queryKey: storyKeys.keyResult(workspaceSlug, keyResultId),
    queryFn: () =>
      getStories({ session: session!, workspaceSlug }, { keyResultId }),
    enabled: Boolean(session && keyResultId && enabled),
  });
};
