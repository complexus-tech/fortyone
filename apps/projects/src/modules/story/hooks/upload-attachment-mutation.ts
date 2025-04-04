import { useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { storyKeys } from "@/modules/stories/constants";
import { addAttachmentAction } from "../actions/add-attachment";
import type { StoryAttachment } from "../types";

export const useUploadAttachmentMutation = (storyId: string) => {
  const queryClient = useQueryClient();
  const toastid = "upload-attachment";

  const mutation = useMutation({
    mutationFn: (file: File) => addAttachmentAction(storyId, file),
    onMutate: async (file) => {
      toast.loading("Uploading...", { id: toastid, description: file.name });
      await queryClient.cancelQueries({
        queryKey: storyKeys.attachments(storyId),
      });
      const previousAttachments =
        queryClient.getQueryData<StoryAttachment[]>(
          storyKeys.attachments(storyId),
        ) || [];
      const optimisticAttachment: StoryAttachment = {
        id: `temp-${Date.now()}`,
        filename: file.name,
        size: file.size,
        mimeType: file.type,
        url: URL.createObjectURL(file),
        createdAt: new Date().toISOString(),
      };
      queryClient.setQueryData(storyKeys.attachments(storyId), [
        ...previousAttachments,
        optimisticAttachment,
      ]);
      return { previousAttachments };
    },

    onError: (error, file, context) => {
      if (context) {
        queryClient.setQueryData(
          storyKeys.attachments(storyId),
          context.previousAttachments,
        );
      }

      toast.error(`Failed to upload attachment: ${file.name}`, {
        id: toastid,
        description:
          error.message || "An error occurred while uploading the file",
        action: {
          label: "Retry",
          onClick: () => {
            mutation.mutate(file);
          },
        },
      });
    },

    onSuccess: (res, _, context) => {
      if (res.error?.message) {
        throw new Error(res.error.message);
      }

      const attachment = res.data!;
      // remove the temp attachment and add the new attachment
      queryClient.setQueryData(storyKeys.attachments(storyId), [
        ...context.previousAttachments,
        attachment,
      ]);

      toast.success("File uploaded", {
        id: toastid,
        description: `${res.data?.filename} uploaded`,
      });
    },
  });

  return mutation;
};
