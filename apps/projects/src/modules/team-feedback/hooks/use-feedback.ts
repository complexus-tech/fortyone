import { useQuery } from "@tanstack/react-query";
import { useSession } from "@/lib/auth/client";
import { useWorkspacePath } from "@/hooks";
import { feedbackKeys } from "@/constants/keys";
import { getTeamFeedbackItem } from "../queries/get-feedback";

export const useTeamFeedbackItem = (feedbackId: string) => {
  const { data: session } = useSession();
  const { workspaceSlug } = useWorkspacePath();

  return useQuery({
    queryKey: feedbackKeys.detail(workspaceSlug, feedbackId),
    queryFn: () =>
      getTeamFeedbackItem(feedbackId, {
        session: session!,
        workspaceSlug,
      }),
    enabled: Boolean(feedbackId && session),
  });
};
