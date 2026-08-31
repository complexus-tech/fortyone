import { useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { useWorkspacePath } from "@/hooks";
import { aiChatKeys } from "../constants";
import { deleteAiChatAction } from "../actions/delete-ai-chat";
import type { AiChatSession } from "../types";

export const useDeleteAiChat = () => {
  const queryClient = useQueryClient();
  const { workspaceSlug } = useWorkspacePath();

  const mutation = useMutation({
    mutationFn: (id: string) => deleteAiChatAction(id, workspaceSlug),
    onMutate: async (id) => {
      await queryClient.cancelQueries({
        queryKey: aiChatKeys.lists(workspaceSlug),
      });
      const previousChats = queryClient.getQueryData<AiChatSession[]>(
        aiChatKeys.lists(workspaceSlug),
      );

      if (previousChats) {
        queryClient.setQueryData<AiChatSession[]>(
          aiChatKeys.lists(workspaceSlug),
          previousChats.filter((chat) => chat.id !== id),
        );
      }

      return { previousChats };
    },
    onError: (error, id, context) => {
      if (context?.previousChats) {
        queryClient.setQueryData(
          aiChatKeys.lists(workspaceSlug),
          context.previousChats,
        );
      }
      toast.error("Failed to delete chat", {
        description: error.message || "Your changes were not saved",
        action: {
          label: "Retry",
          onClick: () => {
            mutation.mutate(id);
          },
        },
      });
    },
    onSuccess: (res, id) => {
      if (res.error?.message) {
        throw new Error(res.error.message);
      }

      queryClient.invalidateQueries({
        queryKey: aiChatKeys.lists(workspaceSlug),
      });
      queryClient.removeQueries({
        queryKey: aiChatKeys.detail(workspaceSlug, id),
      });
      queryClient.removeQueries({
        queryKey: aiChatKeys.messages(workspaceSlug, id),
      });
    },
  });

  return mutation;
};
