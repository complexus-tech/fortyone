import { useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { storyKeys } from "@/modules/stories/constants";
import { addAssociationAction } from "../actions/add-association";
import type { StoryAssociationType } from "../types";

export const useAddAssociationMutation = () => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({
      fromStoryId,
      toStoryId,
      type,
    }: {
      fromStoryId: string;
      toStoryId: string;
      type: StoryAssociationType;
    }) => addAssociationAction(fromStoryId, { toStoryId, type }),

    onSuccess: (res, { fromStoryId, toStoryId }) => {
      if (res.error) {
        throw new Error(res.error.message);
      }
      queryClient.invalidateQueries({ queryKey: storyKeys.detail(fromStoryId) });
      queryClient.invalidateQueries({ queryKey: storyKeys.detail(toStoryId) });
      toast.success("Association added");
    },

    onError: (error) => {
      toast.error("Failed to add association");
    },
  });
};
