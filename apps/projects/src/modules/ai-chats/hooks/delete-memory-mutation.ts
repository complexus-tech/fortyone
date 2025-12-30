import { useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { aiChatKeys } from "../constants";
import { deleteMemoryAction } from "../actions/delete-memory";
import type { Memory } from "../types";

export const useDeleteMemory = () => {
  const queryClient = useQueryClient();

  const mutation = useMutation({
    mutationFn: (id: string) => deleteMemoryAction(id),
    onMutate: async (id) => {
      await queryClient.cancelQueries({ queryKey: aiChatKeys.memories() });

      const previousMemories = queryClient.getQueryData<Memory[]>(
        aiChatKeys.memories(),
      );

      if (previousMemories) {
        queryClient.setQueryData<Memory[]>(
          aiChatKeys.memories(),
          previousMemories.filter((memory) => memory.id !== id),
        );
      }

      return { previousMemories };
    },
    onError: (error, id, context) => {
      if (context?.previousMemories) {
        queryClient.setQueryData(
          aiChatKeys.memories(),
          context.previousMemories,
        );
      }
      toast.error("Failed to delete memory", {
        description: error.message || "Your memory was not deleted",
        action: {
          label: "Retry",
          onClick: () => mutation.mutate(id),
        },
      });
    },
    onSuccess: (res) => {
      if (res.error?.message) {
        throw new Error(res.error.message);
      }

      queryClient.invalidateQueries({ queryKey: aiChatKeys.memories() });
      toast.success("Memory deleted successfully");
    },
  });

  return mutation;
};
