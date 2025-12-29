import { useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { aiChatKeys } from "../constants";
import { updateMemoryAction } from "../actions/update-memory";
import type { UpdateMemoryPayload, Memory } from "../types";

export const useUpdateMemory = () => {
  const queryClient = useQueryClient();

  const mutation = useMutation({
    mutationFn: (payload: UpdateMemoryPayload) => updateMemoryAction(payload),
    onMutate: async (payload) => {
      await queryClient.cancelQueries({ queryKey: aiChatKeys.memory() });

      const previousMemory = queryClient.getQueryData<Memory>(
        aiChatKeys.memory(),
      );

      if (previousMemory) {
        queryClient.setQueryData<Memory>(aiChatKeys.memory(), {
          ...previousMemory,
          memory: payload.memory,
          updatedAt: new Date().toISOString(),
        });
      }

      return { previousMemory };
    },
    onError: (error, payload, context) => {
      if (context?.previousMemory) {
        queryClient.setQueryData(aiChatKeys.memory(), context.previousMemory);
      }
      toast.error("Failed to update memory", {
        description: error.message || "Your changes were not saved",
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

      queryClient.invalidateQueries({ queryKey: aiChatKeys.memory() });
      toast.success("Memory updated successfully");
    },
  });

  return mutation;
};
