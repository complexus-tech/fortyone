import { useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { redirect } from "next/navigation";
import { invitationKeys, workspaceKeys } from "@/constants/keys";
import { acceptInvitation } from "../actions/accept-invitation";
import type { Invitation } from "../types";

const isFortyOneApp = process.env.NEXT_PUBLIC_DOMAIN === "fortyone.app";

export const useAcceptInvitationMutation = () => {
  const queryClient = useQueryClient();
  const toastId = "accept-invitation";

  const mutation = useMutation({
    mutationFn: (inviteToken: string) => acceptInvitation(inviteToken),
    onMutate: async (inviteToken) => {
      await queryClient.cancelQueries({
        queryKey: invitationKeys.mine,
      });

      const previousMineInvitations = queryClient.getQueryData<Invitation[]>(
        invitationKeys.mine,
      );

      // get invitation that is being accepted
      const invitation = previousMineInvitations?.find(
        (invitation) => invitation.token === inviteToken,
      );

      // Optimistically remove the invitation from the cache
      queryClient.setQueryData(invitationKeys.mine, (old: Invitation[] = []) =>
        old.filter((invitation) => invitation.token !== inviteToken),
      );

      toast.loading("Accepting invitation...", {
        description: "Please wait...",
        id: toastId,
      });

      return { previousMineInvitations, invitation };
    },
    onError: (error, variables, context) => {
      // Restore the previous data
      queryClient.setQueryData(
        invitationKeys.mine,
        context?.previousMineInvitations,
      );

      toast.error("Failed to accept", {
        id: toastId,
        description: error.message || "Failed to accept invitation",
        action: {
          label: "Retry",
          onClick: () => {
            mutation.mutate(variables);
          },
        },
      });
    },
    onSuccess: (res, __, context) => {
      if (res.error?.message) {
        throw new Error(res.error.message);
      }

      if (context.invitation) {
        toast.success("Accepted", {
          description: "Invitation accepted successfully",
          id: toastId,
          action: {
            label: "Open",
            onClick: () => {
              if (isFortyOneApp) {
                redirect(`https://${context.invitation!.workspaceSlug}.fortyone.app/my-work`);
              } else {
                redirect(`/${context.invitation!.workspaceSlug}/my-work`);
              }
            },
          },
        });
      } else {
        toast.success("Accepted", {
          description: "Invitation accepted successfully",
          id: toastId,
        });
      }
    },
    onSettled: () => {
      queryClient.invalidateQueries({ queryKey: invitationKeys.mine });
      queryClient.invalidateQueries({
        queryKey: workspaceKeys.lists(),
      });
    },
  });

  return mutation;
};
