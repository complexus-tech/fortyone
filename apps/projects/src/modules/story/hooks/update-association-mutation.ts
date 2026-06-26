import { useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { useWorkspacePath } from "@/hooks";
import { storyKeys } from "@/modules/stories/constants";
import type { DetailedStory, StoryAssociationType } from "../types";
import { updateAssociationAction } from "../actions/update-association";

type UpdateAssociationVariables = {
  associationId: string;
  fromStoryId: string;
  storyId: string;
  toStoryId: string;
  type: StoryAssociationType;
};

type UpdateAssociationContext = {
  previousStory?: DetailedStory;
};

export const useUpdateAssociationMutation = () => {
  const queryClient = useQueryClient();
  const { workspaceSlug } = useWorkspacePath();

  return useMutation<
    Awaited<ReturnType<typeof updateAssociationAction>>,
    Error,
    UpdateAssociationVariables,
    UpdateAssociationContext
  >({
    mutationFn: async ({ associationId, fromStoryId, toStoryId, type }) => {
      const response = await updateAssociationAction(
        associationId,
        {
          fromStoryId,
          toStoryId,
          type,
        },
        workspaceSlug,
      );
      if (response.error?.message) {
        throw new Error(response.error.message);
      }
      return response;
    },

    onMutate: async ({
      associationId,
      fromStoryId,
      storyId,
      toStoryId,
      type,
    }) => {
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
            associations: previousStory.associations.map((association) =>
              association.id === associationId
                ? {
                    ...association,
                    fromStoryId,
                    toStoryId,
                    type,
                  }
                : association,
            ),
          },
        );
      }

      return { previousStory };
    },

    onError: (_error, { storyId }, context) => {
      if (context?.previousStory) {
        queryClient.setQueryData(
          storyKeys.detail(workspaceSlug, storyId),
          context.previousStory,
        );
      }
      toast.error("Failed to update association");
    },

    onSuccess: (_response, { storyId }) => {
      queryClient.invalidateQueries({
        queryKey: storyKeys.detail(workspaceSlug, storyId),
      });
      toast.success("Association updated");
    },
  });
};
