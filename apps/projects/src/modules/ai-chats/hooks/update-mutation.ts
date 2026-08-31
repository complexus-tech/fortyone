import { useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { useWorkspacePath } from "@/hooks";
import { aiChatKeys } from "../constants";
import { updateAiChatAction } from "../actions/update-ai-chat";
import type { UpdateAiChatPayload, AiChatSession } from "../types";

export const useUpdateAiChat = () => {
  const queryClient = useQueryClient();
  const { workspaceSlug } = useWorkspacePath();

  const mutation = useMutation({
    mutationFn: ({
      id,
      payload,
    }: {
      id: string;
      payload: UpdateAiChatPayload;
    }) => updateAiChatAction(id, payload, workspaceSlug),
    onMutate: async ({ id, payload }) => {
      await queryClient.cancelQueries({
        queryKey: aiChatKeys.lists(workspaceSlug),
      });
      await queryClient.cancelQueries({
        queryKey: aiChatKeys.detail(workspaceSlug, id),
      });

      const previousChats = queryClient.getQueryData<AiChatSession[]>(
        aiChatKeys.lists(workspaceSlug),
      );
      const previousChat = queryClient.getQueryData<AiChatSession>(
        aiChatKeys.detail(workspaceSlug, id),
      );

      if (previousChats) {
        queryClient.setQueryData<AiChatSession[]>(
          aiChatKeys.lists(workspaceSlug),
          previousChats.map((chat) =>
            chat.id === id ? { ...chat, ...payload } : chat,
          ),
        );
      }

      if (previousChat) {
        queryClient.setQueryData<AiChatSession>(
          aiChatKeys.detail(workspaceSlug, id),
          {
            ...previousChat,
            ...payload,
          },
        );
      }

      return { previousChats, previousChat };
    },
    onError: (error, { id }, context) => {
      if (context?.previousChats) {
        queryClient.setQueryData(
          aiChatKeys.lists(workspaceSlug),
          context.previousChats,
        );
      }
      if (context?.previousChat) {
        queryClient.setQueryData(
          aiChatKeys.detail(workspaceSlug, id),
          context.previousChat,
        );
      }
      toast.error("Failed to update chat", {
        description: error.message || "Your changes were not saved",
        action: {
          label: "Retry",
          onClick: () => {
            mutation.mutate({ id, payload: { title: "" } });
          },
        },
      });
    },
    onSuccess: (res, { id }) => {
      if (res.error?.message) {
        throw new Error(res.error.message);
      }

      queryClient.invalidateQueries({
        queryKey: aiChatKeys.lists(workspaceSlug),
      });
      queryClient.invalidateQueries({
        queryKey: aiChatKeys.detail(workspaceSlug, id),
      });
    },
  });

  return mutation;
};
