import { useMutation } from "@tanstack/react-query";
import { toast } from "sonner";
import { useWorkspacePath } from "@/hooks";
import { createGitHubInstallSessionAction } from "@/lib/actions/github/create-install-session";

export const useCreateGitHubInstallSession = () => {
  const { workspaceSlug } = useWorkspacePath();

  return useMutation({
    mutationFn: () => createGitHubInstallSessionAction(workspaceSlug),
    onSuccess: (res) => {
      if (res.error?.message) {
        toast.error("GitHub", { description: res.error.message });
        return;
      }
      const installUrl = res.data?.installUrl;
      if (installUrl) {
        window.location.href = installUrl;
      }
    },
  });
};
