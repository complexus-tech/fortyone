import { useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { useWorkspacePath } from "@/hooks";
import { integrationRequestKeys, storyKeys } from "@/constants/keys";
import { acceptIntegrationRequestAction } from "../actions/accept";

export const useAcceptIntegrationRequest = () => {
  const queryClient = useQueryClient();
  const { workspaceSlug } = useWorkspacePath();

  return useMutation({
    mutationFn: (requestId: string) =>
      acceptIntegrationRequestAction(requestId, workspaceSlug),
    onSuccess: (res) => {
      if (res.error?.message) {
        toast.error("Request", { description: res.error.message });
        return;
      }
      toast.success("Request accepted");
      if (res.data?.acceptedStoryId) {
        queryClient.invalidateQueries({
          queryKey: storyKeys.detail(workspaceSlug, res.data.acceptedStoryId),
        });
      }
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
