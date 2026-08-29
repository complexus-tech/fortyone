import { useQuery } from "@tanstack/react-query";
import { useSession } from "@/lib/auth/client";
import { useWorkspacePath } from "@/hooks";
import { feedbackKeys } from "@/constants/keys";
import {
  getTeamFeedbackItem,
  getTeamFeedbackPrivateAuthor,
} from "../queries/get-feedback";

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

export const useTeamFeedbackPrivateAuthor = (
  feedbackId: string,
  enabled: boolean,
) => {
  const { data: session } = useSession();
  const { workspaceSlug } = useWorkspacePath();

  return useQuery({
    queryKey: feedbackKeys.privateAuthor(workspaceSlug, feedbackId),
    queryFn: () =>
      getTeamFeedbackPrivateAuthor(feedbackId, {
        session: session!,
        workspaceSlug,
      }),
    enabled: Boolean(enabled && feedbackId && session),
  });
};
