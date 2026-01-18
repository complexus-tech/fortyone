import { useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { useWorkspacePath } from "@/hooks";
import { aiChatKeys } from "../constants";
import { createMemoryAction } from "../actions/create-memory";
import type { CreateMemoryPayload, Memory } from "../types";

export const useCreateMemory = () => {
  const queryClient = useQueryClient();
  const { workspaceSlug } = useWorkspacePath();

  const mutation = useMutation({
    mutationFn: (payload: CreateMemoryPayload) =>
      createMemoryAction(payload, workspaceSlug),
    onMutate: async (payload) => {
      await queryClient.cancelQueries({ queryKey: aiChatKeys.memories() });

      const previousMemories = queryClient.getQueryData<Memory[]>(
        aiChatKeys.memories(),
      );

      // Create optimistic memory item
      const optimisticMemory: Memory = {
        id: `temp-${Date.now()}`,
        userId: "",
        workspaceId: "",
        content: payload.content,
        createdAt: new Date().toISOString(),
        updatedAt: new Date().toISOString(),
      };

      if (previousMemories) {
        queryClient.setQueryData<Memory[]>(aiChatKeys.memories(), [
          ...previousMemories,
          optimisticMemory,
        ]);
      } else {
        queryClient.setQueryData<Memory[]>(aiChatKeys.memories(), [
          optimisticMemory,
        ]);
      }

      return { previousMemories };
    },
    onError: (error, payload, context) => {
      if (context?.previousMemories) {
        queryClient.setQueryData(
          aiChatKeys.memories(),
          context.previousMemories,
        );
      }
      toast.error("Failed to create memory", {
        description: error.message || "Your memory was not saved",
        action: {
          label: "Retry",
          onClick: () => mutation.mutate(payload),
        },
      });
    },
    onSuccess: (res) => {
      if (res.error?.message) {
        throw new Error(res.error.message);
      }

      queryClient.invalidateQueries({ queryKey: aiChatKeys.memories() });
      toast.success("Memory created successfully");
    },
  });

  return mutation;
};
