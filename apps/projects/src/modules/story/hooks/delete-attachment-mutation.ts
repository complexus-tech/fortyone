import { useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { useWorkspacePath } from "@/hooks";
import { storyKeys } from "@/modules/stories/constants";
import { getDesktopBridge } from "@/shell/desktop/bridge";
import { useDesktopWorkspaceScope } from "@/shell/desktop/workspace-scope";
import { resolveDesktopOperation } from "@/shell/desktop/desktop-outbox-coordinator";
import { deleteStoryAttachmentAction } from "../actions/delete-attachment";
import type { StoryAttachment } from "../types";

export const useDeleteAttachmentMutation = (storyId: string) => {
  const queryClient = useQueryClient();
  const toastId = "delete-attachment";
  const { workspaceSlug } = useWorkspacePath();
  const scope = useDesktopWorkspaceScope();
  const desktop = Boolean(getDesktopBridge()?.readOutbox);

  const mutation = useMutation({
    networkMode: desktop ? "always" : "online",
    mutationFn: async (attachmentId: string) => {
      const read = getDesktopBridge()?.readOutbox;
      if (read) {
        if (!scope)
          throw new Error("Wait for this workspace to finish loading.");
        const snapshot = await read(scope);
        const upload = snapshot.operations.find(
          (operation) =>
            operation.kind === "attachment.upload" &&
            operation.resourceId === storyId &&
            operation.payload.attachmentId === attachmentId,
        );
        if (upload) {
          await resolveDesktopOperation(scope, {
            operationId: upload.id,
            resolution: "discard",
          });
          return { local: true };
        }
        if (!navigator.onLine) {
          throw new Error(
            "Connect to remove an attachment already uploaded to the server.",
          );
        }
      }
      const response = await deleteStoryAttachmentAction(
        storyId,
        attachmentId,
        workspaceSlug,
      );
      if (response.error?.message) throw new Error(response.error.message);
      return { local: false };
    },

    onMutate: async (attachmentId) => {
      toast.loading("Deleting attachment...", { id: toastId });

      await queryClient.cancelQueries({
        queryKey: storyKeys.attachments(workspaceSlug, storyId),
      });

      const previousAttachments =
        queryClient.getQueryData<StoryAttachment[]>(
          storyKeys.attachments(workspaceSlug, storyId),
        ) || [];

      queryClient.setQueryData(
        storyKeys.attachments(workspaceSlug, storyId),
        previousAttachments.filter(
          (attachment) => attachment.id !== attachmentId,
        ),
      );

      return { previousAttachments };
    },

    onError: (error, attachmentId, context) => {
      if (context) {
        queryClient.setQueryData(
          storyKeys.attachments(workspaceSlug, storyId),
          context.previousAttachments,
        );
      }

      toast.error("Failed to delete attachment", {
        id: toastId,
        description: error.message || "Your changes were not saved",
        action: {
          label: "Retry",
          onClick: () => {
            mutation.mutate(attachmentId);
          },
        },
      });
    },
    onSuccess: (res) => {
      if (!res.local) {
        void queryClient.invalidateQueries({
          queryKey: storyKeys.attachments(workspaceSlug, storyId),
        });
      }
      toast.success("Attachment deleted", {
        id: toastId,
        action: null,
      });
    },
  });

  return mutation;
};
