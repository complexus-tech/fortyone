import { useMutation, useQueryClient } from "@tanstack/react-query";
import { slackKeys } from "@/constants/keys";
import { useWorkspacePath } from "@/hooks";
import { createSlackAccountLinkSessionAction } from "@/lib/actions/slack/create-account-link-session";

export const useCreateSlackAccountLinkSession = () => {
  const queryClient = useQueryClient();
  const { workspaceSlug } = useWorkspacePath();

  return useMutation({
    mutationFn: (returnUrl: string) =>
      createSlackAccountLinkSessionAction(workspaceSlug, returnUrl),
    onSuccess: (response) => {
      if (response.data?.linked) {
        void queryClient.invalidateQueries({
          queryKey: slackKeys.integration(workspaceSlug),
        });
      }
    },
  });
};
