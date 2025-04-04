import { useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { storyKeys } from "@/modules/stories/constants";
import { deleteStoryAttachmentAction } from "../actions/delete-attachment";
import type { StoryAttachment } from "../types";

export const useDeleteAttachmentMutation = (storyId: string) => {
  const queryClient = useQueryClient();
  const toastId = "delete-attachment";

  const mutation = useMutation({
    mutationFn: (attachmentId: string) =>
      deleteStoryAttachmentAction(storyId, attachmentId),

    onMutate: async (attachmentId) => {
      toast.loading("Deleting attachment...", { id: toastId });

      await queryClient.cancelQueries({
        queryKey: storyKeys.attachments(storyId),
      });

      const previousAttachments =
        queryClient.getQueryData<StoryAttachment[]>(
          storyKeys.attachments(storyId),
        ) || [];

      queryClient.setQueryData(
        storyKeys.attachments(storyId),
        previousAttachments.filter(
          (attachment) => attachment.id !== attachmentId,
        ),
      );

      return { previousAttachments };
    },

    onError: (error, attachmentId, context) => {
      if (context) {
        queryClient.setQueryData(
          storyKeys.attachments(storyId),
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
      if (res.error?.message) {
        throw new Error(res.error.message);
      }
      queryClient.invalidateQueries({
        queryKey: storyKeys.attachments(storyId),
      });
      toast.success("Attachment deleted", {
        id: toastId,
      });
    },
  });

  return mutation;
};
