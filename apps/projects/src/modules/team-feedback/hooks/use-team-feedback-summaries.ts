import { useQuery } from "@tanstack/react-query";
import { feedbackKeys } from "@/constants/keys";
import { useWorkspacePath } from "@/hooks";
import { useSession } from "@/lib/auth/client";
import { getTeamFeedbackSummaries } from "../queries/get-team-feedback-summaries";

export const useTeamFeedbackSummaries = () => {
  const { data: session } = useSession();
  const { workspaceSlug } = useWorkspacePath();

  return useQuery({
    queryKey: feedbackKeys.teamSummaries(workspaceSlug),
    queryFn: () =>
      getTeamFeedbackSummaries({ session: session!, workspaceSlug }),
    enabled: Boolean(session),
  });
};
