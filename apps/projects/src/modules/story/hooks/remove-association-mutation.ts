import { useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { storyKeys } from "@/modules/stories/constants";
import type { DetailedStory } from "../types";
import { removeAssociationAction } from "../actions/remove-association";

export const useRemoveAssociationMutation = () => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({
      associationId,
    }: {
      associationId: string;
      storyId: string;
    }) => removeAssociationAction(associationId),

    onMutate: async ({ storyId, associationId }) => {
      await queryClient.cancelQueries({ queryKey: storyKeys.detail(storyId) });
      const previousStory = queryClient.getQueryData<DetailedStory>(
        storyKeys.detail(storyId),
      );

      if (previousStory) {
        queryClient.setQueryData<DetailedStory>(storyKeys.detail(storyId), {
          ...previousStory,
          associations: previousStory.associations.filter(
            (a) => a.id !== associationId,
          ),
        });
      }

      return { previousStory };
    },

    onError: (error, { storyId }, context) => {
      if (context?.previousStory) {
        queryClient.setQueryData(
          storyKeys.detail(storyId),
          context.previousStory,
        );
      }
      toast.error("Failed to remove association");
    },

    onSuccess: (res, { storyId }) => {
      if (res.error) {
        throw new Error(res.error.message);
      }
      queryClient.invalidateQueries({ queryKey: storyKeys.detail(storyId) });
      toast.success("Association removed");
    },
  });
};
