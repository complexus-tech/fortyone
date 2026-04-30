import { useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { useWorkspacePath } from "@/hooks";
import { integrationRequestKeys } from "@/constants/keys";
import { declineIntegrationRequestAction } from "../actions/decline";

export const useDeclineIntegrationRequest = () => {
  const queryClient = useQueryClient();
  const { workspaceSlug } = useWorkspacePath();

  return useMutation({
    mutationFn: (requestId: string) =>
      declineIntegrationRequestAction(requestId, workspaceSlug),
    onSuccess: (res) => {
      if (res.error?.message) {
        toast.error("Request", { description: res.error.message });
        return;
      }
      toast.success("Request declined");
    },
    onSettled: () => {
      queryClient.invalidateQueries({
        queryKey: integrationRequestKeys.all(workspaceSlug),
      });
    },
  });
};
