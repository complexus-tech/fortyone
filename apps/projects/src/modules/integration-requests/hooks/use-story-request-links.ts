import { useQuery } from "@tanstack/react-query";
import { integrationRequestKeys } from "@/constants/keys";
import { useWorkspacePath } from "@/hooks";
import { useSession } from "@/lib/auth/client";
import { getStoryIntegrationRequestLinks } from "../queries/get-story-request-links";

export const useStoryIntegrationRequestLinks = (storyId: string) => {
  const { data: session } = useSession();
  const { workspaceSlug } = useWorkspacePath();
  return useQuery({
    queryKey: integrationRequestKeys.storyLinks(workspaceSlug, storyId),
    queryFn: () =>
      getStoryIntegrationRequestLinks(storyId, {
        session: session!,
        workspaceSlug,
      }),
    enabled: Boolean(storyId && session),
  });
};
