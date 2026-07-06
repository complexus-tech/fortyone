import { useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { slackKeys } from "@/constants/keys";
import { useWorkspacePath } from "@/hooks";
import { disconnectSlackWorkspaceAction } from "@/lib/actions/slack/disconnect-workspace";

export const useDisconnectSlackWorkspace = () => {
  const queryClient = useQueryClient();
  const { workspaceSlug } = useWorkspacePath();
  const queryKey = slackKeys.integration(workspaceSlug);

  return useMutation({
    mutationFn: () => disconnectSlackWorkspaceAction(workspaceSlug),
    onSuccess: (res) => {
      if (res.error?.message) {
        toast.error("Slack", { description: res.error.message });
        return;
      }
      toast.success("Slack workspace disconnected");
      queryClient.invalidateQueries({ queryKey });
    },
  });
};
