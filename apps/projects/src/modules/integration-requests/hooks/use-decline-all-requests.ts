import { useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { useWorkspacePath } from "@/hooks";
import { integrationRequestKeys } from "@/constants/keys";
import { declineAllIntegrationRequestsAction } from "../actions/decline-all";

export const useDeclineAllIntegrationRequests = () => {
  const queryClient = useQueryClient();
  const { workspaceSlug } = useWorkspacePath();

  return useMutation({
    mutationFn: (teamId: string) =>
      declineAllIntegrationRequestsAction(teamId, workspaceSlug),
    onSuccess: (res) => {
      if (res.error?.message) {
        toast.error("Intake", { description: res.error.message });
        return;
      }
      toast.success("Intake items declined", {
        description: `${res.data?.count ?? 0} intake item${res.data?.count === 1 ? "" : "s"} declined`,
      });
    },
    onSettled: () => {
      queryClient.invalidateQueries({
        queryKey: integrationRequestKeys.all(workspaceSlug),
      });
    },
  });
};
