import { useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { useWorkspacePath } from "@/hooks";
import { aiChatKeys } from "../constants";
import { updateMemoryAction } from "../actions/update-memory";
import type { UpdateMemoryPayload, Memory } from "../types";

export const useUpdateMemory = () => {
  const queryClient = useQueryClient();
  const { workspaceSlug } = useWorkspacePath();

  const mutation = useMutation({
    mutationFn: ({
      id,
      payload,
    }: {
      id: string;
      payload: UpdateMemoryPayload;
    }) => updateMemoryAction(id, payload, workspaceSlug),
    onMutate: async ({ id, payload }) => {
      await queryClient.cancelQueries({ queryKey: aiChatKeys.memories() });

      const previousMemories = queryClient.getQueryData<Memory[]>(
        aiChatKeys.memories(),
      );

      if (previousMemories) {
        queryClient.setQueryData<Memory[]>(
          aiChatKeys.memories(),
          previousMemories.map((memory) =>
            memory.id === id
              ? {
                  ...memory,
                  content: payload.content,
                  updatedAt: new Date().toISOString(),
                }
              : memory,
          ),
        );
      }

      return { previousMemories };
    },
    onError: (error, { id, payload }, context) => {
      if (context?.previousMemories) {
        queryClient.setQueryData(
          aiChatKeys.memories(),
          context.previousMemories,
        );
      }
      toast.error("Failed to update memory", {
        description: error.message || "Your changes were not saved",
        action: {
          label: "Retry",
          onClick: () => mutation.mutate({ id, payload }),
        },
      });
    },
    onSuccess: (res) => {
      if (res.error?.message) {
        throw new Error(res.error.message);
      }

      queryClient.invalidateQueries({ queryKey: aiChatKeys.memories() });
      toast.success("Memory updated successfully");
    },
  });

  return mutation;
};
