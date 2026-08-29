import { useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { useWorkspacePath } from "@/hooks";
import { storyKeys } from "@/modules/stories/constants";
import type { Story } from "@/modules/stories/types";
import { addAssociationAction } from "../actions/add-association";
import type {
  DetailedStory,
  StoryAssociation,
  StoryAssociationType,
} from "../types";

type AddAssociationVariables = {
  associatedStory: Story;
  fromStoryId: string;
  storyId: string;
  toStoryId: string;
  type: StoryAssociationType;
};

type AddAssociationContext = {
  optimisticAssociationId: string;
  previousAssociations?: StoryAssociation[];
};

const getOptimisticAssociationId = (fromStoryId: string, toStoryId: string) =>
  `optimistic-association-${fromStoryId}-${toStoryId}`;

export const useAddAssociationMutation = () => {
  const queryClient = useQueryClient();
  const { workspaceSlug } = useWorkspacePath();

  return useMutation<
    Awaited<ReturnType<typeof addAssociationAction>>,
    Error,
    AddAssociationVariables,
    AddAssociationContext
  >({
    mutationFn: async ({ fromStoryId, toStoryId, type }) => {
      const response = await addAssociationAction(
        fromStoryId,
        { toStoryId, type },
        workspaceSlug,
      );
      if (response.error?.message) {
        throw new Error(response.error.message);
      }
      return response;
    },

    onMutate: async ({
      associatedStory,
      fromStoryId,
      storyId,
      toStoryId,
      type,
    }) => {
      const storyKey = storyKeys.detail(workspaceSlug, storyId);
      await queryClient.cancelQueries({ queryKey: storyKey });

      const previousStory = queryClient.getQueryData<DetailedStory>(storyKey);
      const optimisticAssociationId = getOptimisticAssociationId(
        fromStoryId,
        toStoryId,
      );

      if (
        previousStory &&
        !previousStory.associations.some(
          (association) => association.story.id === associatedStory.id,
        )
      ) {
        const optimisticAssociation: StoryAssociation = {
          fromStoryId,
          id: optimisticAssociationId,
          story: associatedStory,
          toStoryId,
          type,
        };
        queryClient.setQueryData<DetailedStory>(storyKey, {
          ...previousStory,
          associations: [...previousStory.associations, optimisticAssociation],
        });
      }

      return {
        optimisticAssociationId,
        previousAssociations: previousStory?.associations,
      };
    },

    onError: (_error, { storyId }, context) => {
      const previousAssociations = context?.previousAssociations;
      if (previousAssociations) {
        queryClient.setQueryData<DetailedStory>(
          storyKeys.detail(workspaceSlug, storyId),
          (currentStory) =>
            currentStory
              ? {
                  ...currentStory,
                  associations: previousAssociations,
                }
              : currentStory,
        );
      }
      toast.error("Failed to add association");
    },

    onSuccess: (
      response,
      { associatedStory, fromStoryId, storyId, toStoryId, type },
      context,
    ) => {
      const persistedAssociation = response.data;
      if (persistedAssociation) {
        queryClient.setQueryData<DetailedStory>(
          storyKeys.detail(workspaceSlug, storyId),
          (currentStory) => {
            if (!currentStory) return currentStory;
            return {
              ...currentStory,
              associations: currentStory.associations.map((association) =>
                association.id === context.optimisticAssociationId
                  ? {
                      ...persistedAssociation,
                      fromStoryId,
                      story: associatedStory,
                      toStoryId,
                      type,
                    }
                  : association,
              ),
            };
          },
        );
      }

      queryClient.invalidateQueries({
        queryKey: storyKeys.detail(workspaceSlug, fromStoryId),
      });
      queryClient.invalidateQueries({
        queryKey: storyKeys.detail(workspaceSlug, toStoryId),
      });
      queryClient.invalidateQueries({
        queryKey: storyKeys.activitiesInfinite(workspaceSlug, fromStoryId),
        refetchType: "all",
      });
      queryClient.invalidateQueries({
        queryKey: storyKeys.activitiesInfinite(workspaceSlug, toStoryId),
        refetchType: "all",
      });
      toast.success("Association added");
    },
  });
};
