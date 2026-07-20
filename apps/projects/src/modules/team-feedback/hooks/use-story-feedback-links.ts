import { useQuery } from "@tanstack/react-query";
import { feedbackKeys } from "@/constants/keys";
import { useWorkspacePath } from "@/hooks";
import { useSession } from "@/lib/auth/client";
import { getStoryFeedbackLinks } from "../queries/get-story-feedback-links";

export const useStoryFeedbackLinks = (storyId: string) => {
  const { data: session } = useSession();
  const { workspaceSlug } = useWorkspacePath();

  return useQuery({
    queryKey: feedbackKeys.storyLinks(workspaceSlug, storyId),
    queryFn: () =>
      getStoryFeedbackLinks(storyId, {
        session: session!,
        workspaceSlug,
      }),
    enabled: Boolean(storyId && session),
  });
};
