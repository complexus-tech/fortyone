import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { feedbackKeys } from "@/constants/keys";
import { useSession } from "@/lib/auth/client";
import { useWorkspacePath } from "@/hooks";
import { DURATION_FROM_MILLISECONDS } from "@/constants/time";
import { createFeedbackBoard, updateFeedbackPortal } from "./actions";
import { getFeedbackPortals } from "./queries";

export const useFeedbackPortals = () => {
  const { data: session } = useSession();
  const { workspaceSlug } = useWorkspacePath();

  return useQuery({
    queryKey: feedbackKeys.portals(workspaceSlug),
    queryFn: () => getFeedbackPortals({ session: session!, workspaceSlug }),
    enabled: Boolean(session),
    staleTime: DURATION_FROM_MILLISECONDS.MINUTE * 5,
  });
};

export const useUpdateFeedbackPortalMutation = () => {
  const queryClient = useQueryClient();
  const { workspaceSlug } = useWorkspacePath();

  return useMutation({
    mutationFn: ({
      portalId,
      input,
    }: {
      portalId: string;
      input: Parameters<typeof updateFeedbackPortal>[1];
    }) => updateFeedbackPortal(portalId, input, workspaceSlug),
    onSuccess: (response) => {
      if (response.error?.message) {
        toast.error("Failed to update portal", {
          description: response.error.message,
        });
        return;
      }
      toast.success("Portal updated");
      void queryClient.invalidateQueries({
        queryKey: feedbackKeys.portals(workspaceSlug),
      });
    },
    onError: (error) => {
      toast.error("Failed to update portal", {
        description: error.message,
      });
    },
  });
};

export const useCreateFeedbackBoardMutation = () => {
  const queryClient = useQueryClient();
  const { workspaceSlug } = useWorkspacePath();

  return useMutation({
    mutationFn: (input: Parameters<typeof createFeedbackBoard>[0]) =>
      createFeedbackBoard(input, workspaceSlug),
    onSuccess: (response) => {
      if (response.error?.message) {
        toast.error("Failed to create board", {
          description: response.error.message,
        });
        return;
      }
      toast.success("Board created");
      void queryClient.invalidateQueries({
        queryKey: feedbackKeys.portals(workspaceSlug),
      });
    },
    onError: (error) => {
      toast.error("Failed to create board", {
        description: error.message,
      });
    },
  });
};
