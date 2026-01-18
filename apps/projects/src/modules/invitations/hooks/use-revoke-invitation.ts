import { useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { invitationKeys } from "@/constants/keys";
import { useWorkspacePath } from "@/hooks";
import { revokeInvitation } from "../actions/revoke";
import type { Invitation } from "../types";

export const useRevokeInvitationMutation = () => {
  const queryClient = useQueryClient();
  const { workspaceSlug } = useWorkspacePath();
  const toastId = "revoke-invitation";

  const mutation = useMutation({
    mutationFn: (invitationId: string) =>
      revokeInvitation(invitationId, workspaceSlug),
    onMutate: async (invitationId) => {
      await queryClient.cancelQueries({
        queryKey: invitationKeys.pending(workspaceSlug),
      });
      await queryClient.cancelQueries({
        queryKey: invitationKeys.mine,
      });

      const previousInvitations = queryClient.getQueryData<Invitation[]>(
        invitationKeys.pending(workspaceSlug),
      );
      const previousMineInvitations = queryClient.getQueryData<Invitation[]>(
        invitationKeys.mine,
      );

      // Optimistically remove the invitation from the cache
      queryClient.setQueryData(
        invitationKeys.pending(workspaceSlug),
        (old: Invitation[] = []) =>
          old.filter((invitation) => invitation.id !== invitationId),
      );
      queryClient.setQueryData(invitationKeys.mine, (old: Invitation[] = []) =>
        old.filter((invitation) => invitation.id !== invitationId),
      );

      toast.loading("Revoking invitation...", {
        description: "Please wait...",
        id: toastId,
      });

      return { previousInvitations, previousMineInvitations };
    },
    onError: (error, variables, context) => {
      // Restore the previous data
      queryClient.setQueryData(
        invitationKeys.pending(workspaceSlug),
        context?.previousInvitations,
      );
      queryClient.setQueryData(
        invitationKeys.mine,
        context?.previousMineInvitations,
      );

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
      queryClient.invalidateQueries({ queryKey: invitationKeys.pending(workspaceSlug) });
      queryClient.invalidateQueries({ queryKey: invitationKeys.mine });
    },
  });

  return mutation;
};
