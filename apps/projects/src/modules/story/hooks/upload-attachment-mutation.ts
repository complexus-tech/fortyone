import { useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { useSession } from "@/lib/auth/client";
import { useWorkspacePath } from "@/hooks";
import { storyKeys } from "@/modules/stories/constants";
import { addAttachmentAction } from "../actions/add-attachment";
import type { StoryAttachment } from "../types";

const createObjectUrlHandle = (file: File) => {
  const url = URL.createObjectURL(file);
  return {
    url,
    revoke: () => {
      URL.revokeObjectURL(url);
    },
  };
};

export const useUploadAttachmentMutation = (storyId: string) => {
  const queryClient = useQueryClient();
  const { data: session } = useSession();
  const { workspaceSlug } = useWorkspacePath();

  const mutation = useMutation({
    mutationFn: async (file: File) => {
      const response = await addAttachmentAction(storyId, file, workspaceSlug);
      if (response.error?.message) {
        throw new Error(response.error.message);
      }
      if (!response.data) {
        throw new Error("The server did not return the uploaded file.");
      }
      return response.data;
    },
    onMutate: async (file) => {
      const optimisticId = `temp-${crypto.randomUUID()}`;
      const preview = createObjectUrlHandle(file);
      const toastId = `upload-attachment-${optimisticId}`;

      toast.loading("Uploading...", { id: toastId, description: file.name });
      await queryClient.cancelQueries({
        queryKey: storyKeys.attachments(workspaceSlug, storyId),
      });
      const optimisticAttachment: StoryAttachment = {
        id: optimisticId,
        filename: file.name,
        size: file.size,
        mimeType: file.type,
        url: preview.url,
        createdAt: new Date().toISOString(),
        uploadedBy: session ? session.user.id : "",
      };
      queryClient.setQueryData<StoryAttachment[]>(
        storyKeys.attachments(workspaceSlug, storyId),
        (current = []) => [...current, optimisticAttachment],
      );

      return { optimisticId, preview, toastId };
    },

    onError: (error, file, context) => {
      if (context) {
        queryClient.setQueryData<StoryAttachment[]>(
          storyKeys.attachments(workspaceSlug, storyId),
          (current = []) =>
            current.filter(
              (attachment) => attachment.id !== context.optimisticId,
            ),
        );
      }

      toast.error(`Failed to upload attachment: ${file.name}`, {
        id: context?.toastId,
        description: error.message || "The upload failed without a response.",
        action: {
          label: "Retry",
          onClick: () => {
            mutation.mutate(file);
          },
        },
      });
    },

    onSuccess: (attachment, _, context) => {
      queryClient.setQueryData<StoryAttachment[]>(
        storyKeys.attachments(workspaceSlug, storyId),
        (current = []) =>
          current.map((item) =>
            item.id === context.optimisticId ? attachment : item,
          ),
      );

      toast.success("File uploaded", {
        id: context.toastId,
        description: `${attachment.filename} uploaded`,
        action: null,
      });
    },
    onSettled: (_data, _error, _file, context) => {
      context?.preview.revoke();
    },
  });

  return mutation;
};
