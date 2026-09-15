import { useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { redirect } from "next/navigation";
import { buildWorkspaceUrl } from "@/utils/workspace-url";
import {
  acceptUserInvitation,
  InvitationUnavailableError,
} from "../actions/accept-user-invitation";
import {
  optimisticallyAcceptInvitation,
  reconcileAcceptedInvitation,
  rollbackAcceptedInvitation,
} from "../mutations/cache";

export const useAcceptInvitationMutation = () => {
  const queryClient = useQueryClient();
  const toastId = "accept-invitation";

  const mutation = useMutation({
    mutationFn: (invitationId: string) => acceptUserInvitation(invitationId),
    onMutate: async (invitationId) => {
      const context = await optimisticallyAcceptInvitation(
        queryClient,
        invitationId,
      );

      toast.loading("Accepting invitation...", {
        description: "Please wait...",
        id: toastId,
      });

      return context;
    },
    onError: (error, variables, context) => {
      if (error instanceof InvitationUnavailableError) {
        toast.info("Invitation no longer available", {
          id: toastId,
          description: error.message,
        });
        return;
      }

      rollbackAcceptedInvitation(queryClient, context);

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
    onSuccess: (_, __, context) => {
      if (context.invitation) {
        toast.success("Accepted", {
          description: "Invitation accepted successfully",
          id: toastId,
          action: {
            label: "Open",
            onClick: () => {
              redirect(buildWorkspaceUrl(context.invitation!.workspaceSlug));
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
      reconcileAcceptedInvitation(queryClient);
    },
  });

  return mutation;
};
