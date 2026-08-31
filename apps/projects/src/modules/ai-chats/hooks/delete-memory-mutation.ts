import { useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { useWorkspacePath } from "@/hooks";
import { aiChatKeys } from "../constants";
import { deleteMemoryAction } from "../actions/delete-memory";
import type { Memory } from "../types";

export const useDeleteMemory = () => {
  const queryClient = useQueryClient();
  const { workspaceSlug } = useWorkspacePath();

  const mutation = useMutation({
    mutationFn: (id: string) => deleteMemoryAction(id, workspaceSlug),
    onMutate: async (id) => {
      await queryClient.cancelQueries({
        queryKey: aiChatKeys.memories(workspaceSlug),
      });

      const previousMemories = queryClient.getQueryData<Memory[]>(
        aiChatKeys.memories(workspaceSlug),
      );

      if (previousMemories) {
        queryClient.setQueryData<Memory[]>(
          aiChatKeys.memories(workspaceSlug),
          previousMemories.filter((memory) => memory.id !== id),
        );
      }

      return { previousMemories };
    },
    onError: (error, id, context) => {
      if (context?.previousMemories) {
        queryClient.setQueryData(
          aiChatKeys.memories(workspaceSlug),
          context.previousMemories,
        );
      }
      toast.error("Failed to delete memory", {
        description: error.message || "Your memory was not deleted",
        action: {
          label: "Retry",
          onClick: () => {
            mutation.mutate(id);
          },
        },
      });
    },
    onSuccess: (res) => {
      if (res.error?.message) {
        throw new Error(res.error.message);
      }

      queryClient.invalidateQueries({
        queryKey: aiChatKeys.memories(workspaceSlug),
      });
      toast.success("Memory deleted successfully");
    },
  });

  return mutation;
};
