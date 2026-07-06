import { useMutation } from "@tanstack/react-query";
import { toast } from "sonner";
import { useWorkspacePath } from "@/hooks";
import { createSlackInstallSessionAction } from "@/lib/actions/slack/create-install-session";

export const useCreateSlackInstallSession = () => {
  const { workspaceSlug } = useWorkspacePath();

  return useMutation({
    mutationFn: () => createSlackInstallSessionAction(workspaceSlug),
    onSuccess: (res) => {
      if (res.error?.message) {
        toast.error("Slack", { description: res.error.message });
        return;
      }
      const installUrl = res.data?.installUrl;
      if (installUrl) {
        window.location.href = installUrl;
      }
    },
  });
};
