import { useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { useWorkspacePath } from "@/hooks";
import { storyKeys } from "@/modules/stories/constants";
import type { DetailedStory } from "../types";
import { removeAssociationAction } from "../actions/remove-association";

export const useRemoveAssociationMutation = () => {
  const queryClient = useQueryClient();
  const { workspaceSlug } = useWorkspacePath();

  return useMutation({
    mutationFn: ({
      associationId,
    }: {
      associationId: string;
      storyId: string;
    }) => removeAssociationAction(associationId, workspaceSlug),

    onMutate: async ({ storyId, associationId }) => {
      await queryClient.cancelQueries({
        queryKey: storyKeys.detail(workspaceSlug, storyId),
      });
      const previousStory = queryClient.getQueryData<DetailedStory>(
        storyKeys.detail(workspaceSlug, storyId),
      );

      if (previousStory) {
        queryClient.setQueryData<DetailedStory>(
          storyKeys.detail(workspaceSlug, storyId),
          {
            ...previousStory,
            associations: previousStory.associations.filter(
              (a) => a.id !== associationId,
            ),
          },
        );
      }

      return { previousStory };
    },

    onError: (error, { storyId }, context) => {
      if (context?.previousStory) {
        queryClient.setQueryData(
          storyKeys.detail(workspaceSlug, storyId),
          context.previousStory,
        );
      }
      toast.error("Failed to remove association");
    },

    onSuccess: (res, { storyId }) => {
      if (res.error) {
        throw new Error(res.error.message);
      }
      queryClient.invalidateQueries({
        queryKey: storyKeys.detail(workspaceSlug, storyId),
      });
      toast.success("Association removed");
    },
  });
};
