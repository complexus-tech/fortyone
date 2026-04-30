import { useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { useWorkspacePath } from "@/hooks";
import { integrationRequestKeys } from "@/constants/keys";
import { updateIntegrationRequestAction } from "../actions/update";
import type {
  IntegrationRequest,
  UpdateIntegrationRequestInput,
} from "../types";

export const useUpdateIntegrationRequest = () => {
  const queryClient = useQueryClient();
  const { workspaceSlug } = useWorkspacePath();

  return useMutation({
    mutationFn: ({
      requestId,
      payload,
    }: {
      requestId: string;
      payload: UpdateIntegrationRequestInput;
    }) => updateIntegrationRequestAction(requestId, payload, workspaceSlug),
    onMutate: async ({ requestId, payload }) => {
      const detailKey = integrationRequestKeys.detail(workspaceSlug, requestId);
      await queryClient.cancelQueries({ queryKey: detailKey });
      const previous = queryClient.getQueryData<IntegrationRequest>(detailKey);
      if (previous) {
        queryClient.setQueryData<IntegrationRequest>(detailKey, {
          ...previous,
          ...payload,
          updatedAt: new Date().toISOString(),
        });
      }
      return { previous };
    },
    onError: (error, { requestId }, context) => {
      if (context?.previous) {
        queryClient.setQueryData(
          integrationRequestKeys.detail(workspaceSlug, requestId),
          context.previous,
        );
      }
      toast.error("Failed to update request", {
        description: error.message || "Your changes were not saved",
      });
    },
    onSuccess: (res) => {
      if (res.error?.message) {
        toast.error("Failed to update request", {
          description: res.error.message,
        });
      }
    },
    onSettled: (_res, _err, { requestId }) => {
      queryClient.invalidateQueries({
        queryKey: integrationRequestKeys.detail(workspaceSlug, requestId),
      });
      queryClient.invalidateQueries({
        queryKey: integrationRequestKeys.lists(workspaceSlug),
      });
    },
  });
};
