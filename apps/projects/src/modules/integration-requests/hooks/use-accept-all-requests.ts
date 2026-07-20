import { useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { useWorkspacePath } from "@/hooks";
import { integrationRequestKeys } from "@/constants/keys";
import { storyKeys } from "@/modules/stories/constants";
import { acceptAllIntegrationRequestsAction } from "../actions/accept-all";

export const useAcceptAllIntegrationRequests = () => {
  const queryClient = useQueryClient();
  const { workspaceSlug } = useWorkspacePath();

  return useMutation({
    mutationFn: (teamId: string) =>
      acceptAllIntegrationRequestsAction(teamId, workspaceSlug),
    onSuccess: (res) => {
      if (res.error?.message) {
        toast.error("Intake", { description: res.error.message });
        return;
      }
      toast.success("Intake items accepted", {
        description: `${res.data?.count ?? 0} intake item${res.data?.count === 1 ? "" : "s"} accepted`,
      });
    },
    onSettled: () => {
      queryClient.invalidateQueries({
        queryKey: integrationRequestKeys.all(workspaceSlug),
      });
      queryClient.invalidateQueries({
        queryKey: storyKeys.all(workspaceSlug),
      });
    },
  });
};
