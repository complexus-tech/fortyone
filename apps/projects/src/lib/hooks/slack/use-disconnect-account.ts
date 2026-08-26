import { useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { slackKeys } from "@/constants/keys";
import { useWorkspacePath } from "@/hooks";
import { disconnectSlackAccountAction } from "@/lib/actions/slack/disconnect-account";

export const useDisconnectSlackAccount = () => {
  const queryClient = useQueryClient();
  const { workspaceSlug } = useWorkspacePath();
  const queryKey = slackKeys.integration(workspaceSlug);

  return useMutation({
    mutationFn: () => disconnectSlackAccountAction(workspaceSlug),
    onSuccess: (res) => {
      if (res.error?.message) {
        toast.error("Slack", { description: res.error.message });
        return;
      }
      toast.success("Slack account disconnected");
      void queryClient.invalidateQueries({ queryKey });
    },
  });
};
