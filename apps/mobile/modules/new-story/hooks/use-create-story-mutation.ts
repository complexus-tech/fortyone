import { useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner-native";
import { storyKeys } from "@/constants/keys";
import { createStory } from "../actions/create-story";

export const useCreateStoryMutation = () => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: createStory,
    onSuccess: (response) => {
      if (response.error?.message) {
        toast.error(response.error.message);
        return;
      }

      queryClient.invalidateQueries({ queryKey: storyKeys.all });
      toast.success("Task created");
    },
    onError: (error) => {
      toast.error("Failed to create task", {
        description:
          error instanceof Error ? error.message : "Your task was not saved",
      });
    },
  });
};
