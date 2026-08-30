import { useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { useWorkspacePath } from "@/hooks/use-workspace-path";
import { revokeInvitation } from "../actions/revoke";
import {
  optimisticallyRevokeInvitation,
  reconcileRevokedInvitation,
  rollbackRevokedInvitation,
} from "../mutations/cache";

export const useRevokeInvitationMutation = () => {
  const queryClient = useQueryClient();
  const { workspaceSlug } = useWorkspacePath();
  const toastId = "revoke-invitation";

  const mutation = useMutation({
    mutationFn: (invitationId: string) =>
      revokeInvitation(invitationId, workspaceSlug),
    onMutate: async (invitationId) => {
      const context = await optimisticallyRevokeInvitation(
        queryClient,
        workspaceSlug,
        invitationId,
      );

      toast.loading("Revoking invitation...", {
        description: "Please wait...",
        id: toastId,
      });

      return context;
    },
    onError: (error, variables, context) => {
      rollbackRevokedInvitation(queryClient, workspaceSlug, context);

      toast.error("Failed to revoke", {
        id: toastId,
        description: error.message || "Failed to revoke invitation",
        action: {
          label: "Retry",
          onClick: () => {
            mutation.mutate(variables);
          },
        },
      });
    },
    onSuccess: (res) => {
      if (res.error?.message) {
        throw new Error(res.error.message);
      }

      toast.success("Revoked", {
        description: "Invitation revoked successfully",
        id: toastId,
      });
      reconcileRevokedInvitation(queryClient, workspaceSlug);
    },
  });

  return mutation;
};
