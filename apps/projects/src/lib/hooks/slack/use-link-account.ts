import { useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { slackKeys } from "@/constants/keys";
import { useWorkspacePath } from "@/hooks";
import { linkSlackAccountAction } from "@/lib/actions/slack/link-account";

export const useLinkSlackAccount = () => {
  const queryClient = useQueryClient();
  const { workspaceSlug } = useWorkspacePath();
  const queryKey = slackKeys.integration(workspaceSlug);

  return useMutation({
    mutationFn: (token: string) =>
      linkSlackAccountAction(workspaceSlug, { token }),
    onSuccess: (res) => {
      if (res.error?.message) {
        toast.error("Slack", { description: res.error.message });
        return;
      }
      toast.success("Slack account connected");
      queryClient.invalidateQueries({ queryKey });
    },
  });
};
