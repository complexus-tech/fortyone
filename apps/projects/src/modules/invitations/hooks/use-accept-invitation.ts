import { useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { redirect } from "next/navigation";
import { buildWorkspaceUrl } from "@/utils/workspace-url";
import { acceptInvitation } from "../actions/accept-invitation";
import {
  optimisticallyAcceptInvitation,
  reconcileAcceptedInvitation,
  rollbackAcceptedInvitation,
} from "../mutations/cache";

export const useAcceptInvitationMutation = () => {
  const queryClient = useQueryClient();
  const toastId = "accept-invitation";

  const mutation = useMutation({
    mutationFn: (inviteToken: string) => acceptInvitation(inviteToken),
    onMutate: async (inviteToken) => {
      const context = await optimisticallyAcceptInvitation(
        queryClient,
        inviteToken,
      );

      toast.loading("Accepting invitation...", {
        description: "Please wait...",
        id: toastId,
      });

      return context;
    },
    onError: (error, variables, context) => {
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
